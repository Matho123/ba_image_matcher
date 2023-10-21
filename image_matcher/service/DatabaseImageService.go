package service

import (
	"errors"
	"fmt"
	"gocv.io/x/gocv"
	"image/color"
	"image_matcher/client"
	"image_matcher/image-handling"
	"log"
	"time"
)

var descriptorMapping = map[image_handling.FeatureImageAnalyzer]string{
	image_handling.SiftImageAnalyzer{}:  "sift_descriptor",
	image_handling.ORBImageAnalyzer{}:   "orb_descriptor",
	image_handling.BRISKImageAnalyzer{}: "brisk_descriptor",
}

const MaxChunkSize = 50

func AnalyzeAndSaveDatabaseImage(rawImages []*image_handling.RawImage) error {
	databaseConnection, err := openDatabaseConnection()
	if err != nil {
		return err
	}
	defer databaseConnection.Close()

	for _, rawImage := range rawImages {

		_, siftDesc, _ := image_handling.ExtractKeypointsAndDescriptors(&rawImage.Data, image_handling.SiftImageAnalyzer{})
		_, orbDesc, _ := image_handling.ExtractKeypointsAndDescriptors(&rawImage.Data, image_handling.ORBImageAnalyzer{})
		_, briskDesc, _ := image_handling.ExtractKeypointsAndDescriptors(&rawImage.Data, image_handling.BRISKImageAnalyzer{})
		pHash, _ := client.GetPHashValue(rawImage.Data)

		err := insertImageIntoDatabaseSet(
			databaseConnection,
			DatabaseSetImage{
				externalReference: rawImage.ExternalReference,
				siftDescriptor:    image_handling.ConvertImageMatToByteArray(siftDesc),
				orbDescriptor:     image_handling.ConvertImageMatToByteArray(orbDesc),
				briskDescriptor:   image_handling.ConvertImageMatToByteArray(briskDesc),
				pHash:             pHash,
			},
		)

		if err != nil {
			return err
		}
	}
	return nil
}

func MatchAgainstDatabaseFeatureBased(
	searchImage image_handling.RawImage,
	analyzer string,
	matcher string,
	similarityThreshold float64,
) ([]string, error, time.Duration, time.Duration) {
	imageAnalyzer, imageMatcher, err := getAnalyzerAndMatcher(analyzer, matcher)
	if err != nil {
		return []string{}, err, 0, 0
	}
	defer imageMatcher.Close()

	databaseConnection, err := openDatabaseConnection()
	if err != nil {
		return []string{}, err, 0, 0
	}
	defer databaseConnection.Close()

	var matchingTime time.Duration

	_, searchImageDescriptor, extractionTime := image_handling.ExtractKeypointsAndDescriptors(&searchImage.Data, imageAnalyzer)
	defer searchImageDescriptor.Close()

	var matchedImages []string

	offset := 0
	for {
		databaseImages, err := retrieveFeatureImageChunk(
			databaseConnection,
			descriptorMapping[imageAnalyzer],
			offset,
			MaxChunkSize+1,
		)
		if err != nil {
			log.Println("Error while retrieving chunk from database images: ", err)
		}

		for _, databaseImage := range databaseImages {
			println("\nComparing to " + databaseImage.externalReference)
			databaseImageDescriptor, err := image_handling.ConvertByteArrayToDescriptorMat(databaseImage.descriptor, analyzer)

			if databaseImageDescriptor == nil || err != nil {
				println("Descriptor was empty")
				continue
			}

			matchingStart := time.Now()
			matches := imageMatcher.FindMatches(&searchImageDescriptor, databaseImageDescriptor)
			matchingTime += time.Since(matchingStart)
			databaseImageDescriptor.Close()

			isMatch, _ := image_handling.DetermineSimilarity(matches, similarityThreshold)
			if isMatch {
				matchedImages = append(matchedImages, databaseImage.externalReference)
			}
		}

		if len(databaseImages) < MaxChunkSize+1 {
			break
		}
		offset += MaxChunkSize
	}
	return matchedImages, nil, extractionTime, matchingTime
}

func MatchImageAgainstDatabasePHash(searchImage image_handling.RawImage, maxHammingDistance int) ([]string, error, time.Duration,
	time.Duration) {
	databaseConnection, err := openDatabaseConnection()
	if err != nil {
		return []string{}, err, 0, 0
	}
	defer databaseConnection.Close()

	var matchingTime time.Duration

	searchImageHash, extractionTime := client.GetPHashValue(searchImage.Data)

	var matchedImages []string

	offset := 0
	for {
		databaseImages, err := retrievePHashImageChunk(databaseConnection, offset, MaxChunkSize+1)
		if err != nil {
			log.Println("Error while retrieving chunk from database images: ", err)
		}
		matchingStart := time.Now()
		for _, databaseImage := range databaseImages {
			if image_handling.HashesAreMatch(searchImageHash, databaseImage.hash, maxHammingDistance) {
				matchedImages = append(matchedImages, databaseImage.externalReference)
			}
		}
		matchingTime += time.Since(matchingStart)

		if len(databaseImages) < MaxChunkSize+1 {
			break
		}
		offset += MaxChunkSize
	}
	return matchedImages, nil, time.Duration(extractionTime * float64(time.Second)), matchingTime
}

func AnalyzeAndMatchTwoImages(
	image1 image_handling.RawImage,
	image2 image_handling.RawImage,
	analyzer string,
	matcher string,
	similarityThreshold float64,
	debug bool,
) (bool, []gocv.KeyPoint, []gocv.KeyPoint, time.Duration, time.Duration, error) {
	if analyzer == image_handling.PHASH {
		hash1, extractionTime1 := client.GetPHashValue(image1.Data)
		hash2, extractionTime2 := client.GetPHashValue(image2.Data)
		extractionTime := time.Duration((extractionTime1 + extractionTime2) * float64(time.Second))

		log.Println(fmt.Sprintf("hash1: %d | hash2: %d", hash1, hash2))

		startTimeMatching := time.Now()
		imagesAreMatch := image_handling.HashesAreMatch(hash1, hash2, 4)
		matchingTime := time.Since(startTimeMatching)

		return imagesAreMatch, []gocv.KeyPoint{}, []gocv.KeyPoint{}, extractionTime, matchingTime, nil
	}

	//for feature-based analyzer
	imageAnalyzer, imageMatcher, err := getAnalyzerAndMatcher(analyzer, matcher)
	if err != nil {
		return false, nil, nil, 0, 0, err
	}
	defer imageMatcher.Close()

	keypoints1, imageDescriptors1, time1 := image_handling.ExtractKeypointsAndDescriptors(&image1.Data, imageAnalyzer)
	defer imageDescriptors1.Close()
	keypoints2, imageDescriptors2, time2 := image_handling.ExtractKeypointsAndDescriptors(&image2.Data, imageAnalyzer)
	defer imageDescriptors2.Close()
	extractionTime := time1 + time2

	log.Println(imageDescriptors1.ToBytes())

	startTimeMatching := time.Now()
	matches := imageMatcher.FindMatches(&imageDescriptors1, &imageDescriptors2)

	imagesAreMatch, bestMatches := image_handling.DetermineSimilarity(matches, similarityThreshold)
	matchingTime := time.Since(startTimeMatching)

	if debug {
		image1Mat := image_handling.ConvertImageToMat(&image1.Data, color.RGBA{R: 255, G: 255, B: 255, A: 255})
		image2Mat := image_handling.ConvertImageToMat(&image2.Data, color.RGBA{R: 255, G: 255, B: 255, A: 255})
		image_handling.DrawMatches(&image1Mat, keypoints1, &image2Mat, keypoints2, *bestMatches)
	}

	return imagesAreMatch, keypoints1, keypoints2, extractionTime, matchingTime, nil
}

func getAnalyzerAndMatcher(analyzer, matcher string) (image_handling.FeatureImageAnalyzer, image_handling.ImageMatcher, error) {
	imageAnalyzer := image_handling.ImageAnalyzerMapping[analyzer]
	imageMatcher := image_handling.ImageMatcherMapping[matcher]

	if imageAnalyzer == nil || imageMatcher == nil {
		return nil, nil, errors.New("invalid option for analyzer or matcher")
	}
	return imageAnalyzer, imageMatcher, nil
}
