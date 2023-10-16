package service

import (
	"errors"
	"fmt"
	"gocv.io/x/gocv"
	"image/color"
	"image_matcher/client"
	"image_matcher/file-handling"
	"image_matcher/image-matching"
	"log"
	"time"
)

var descriptorMapping = map[image_matching.FeatureImageAnalyzer]string{
	image_matching.SiftImageAnalyzer{}:  "sift_descriptor",
	image_matching.ORBImageAnalyzer{}:   "orb_descriptor",
	image_matching.BRISKImageAnalyzer{}: "brisk_descriptor",
}

const MaxChunkSize = 50

func AnalyzeAndSaveDatabaseImage(rawImages []*file_handling.RawImage) error {
	databaseConnection, err := openDatabaseConnection()
	if err != nil {
		return err
	}
	defer databaseConnection.Close()

	for _, rawImage := range rawImages {

		_, siftDesc, _ := image_matching.ExtractKeypointsAndDescriptors(&rawImage.Data, image_matching.SiftImageAnalyzer{})
		_, orbDesc, _ := image_matching.ExtractKeypointsAndDescriptors(&rawImage.Data, image_matching.ORBImageAnalyzer{})
		_, briskDesc, _ := image_matching.ExtractKeypointsAndDescriptors(&rawImage.Data, image_matching.BRISKImageAnalyzer{})
		pHash, _ := client.GetPHashValue(rawImage.Data)

		err := insertImageIntoDatabaseSet(
			databaseConnection,
			DatabaseSetImage{
				externalReference: rawImage.ExternalReference,
				siftDescriptor:    file_handling.ConvertImageMatToByteArray(siftDesc),
				orbDescriptor:     file_handling.ConvertImageMatToByteArray(orbDesc),
				briskDescriptor:   file_handling.ConvertImageMatToByteArray(briskDesc),
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
	searchImage file_handling.RawImage,
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

	_, searchImageDescriptor, extractionTime := image_matching.ExtractKeypointsAndDescriptors(&searchImage.Data, imageAnalyzer)

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
			databaseImageDescriptor := file_handling.ConvertByteArrayToMat(databaseImage.descriptor)
			matchingStart := time.Now()
			matches := imageMatcher.FindMatches(searchImageDescriptor, databaseImageDescriptor)
			matchingTime += time.Since(matchingStart)

			isMatch, _ := image_matching.DetermineSimilarity(matches, similarityThreshold)
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

func MatchImageAgainstDatabasePHash(searchImage file_handling.RawImage, maxHammingDistance int) ([]string, error, time.Duration,
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
			if image_matching.HashesAreMatch(searchImageHash, databaseImage.hash, maxHammingDistance) {
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
	image1 file_handling.RawImage,
	image2 file_handling.RawImage,
	analyzer string,
	matcher string,
	similarityThreshold float64,
	debug bool,
) (bool, []gocv.KeyPoint, []gocv.KeyPoint, time.Duration, time.Duration, error) {
	if analyzer == "phash" {
		hash1, extractionTime1 := client.GetPHashValue(image1.Data)
		hash2, extractionTime2 := client.GetPHashValue(image2.Data)
		extractionTime := time.Duration((extractionTime1 + extractionTime2) * float64(time.Second))

		log.Println(fmt.Sprintf("hash1: %d | hash2: %d", hash1, hash2))

		startTimeMatching := time.Now()
		imagesAreMatch := image_matching.HashesAreMatch(hash1, hash2, 4)
		matchingTime := time.Since(startTimeMatching)

		return imagesAreMatch, []gocv.KeyPoint{}, []gocv.KeyPoint{}, extractionTime, matchingTime, nil
	}

	//for feature-based analyzer
	imageAnalyzer, imageMatcher, err := getAnalyzerAndMatcher(analyzer, matcher)
	if err != nil {
		return false, nil, nil, 0, 0, err
	}
	defer imageMatcher.Close()

	keypoints1, imageDescriptors1, time1 := image_matching.ExtractKeypointsAndDescriptors(&image1.Data, imageAnalyzer)
	keypoints2, imageDescriptors2, time2 := image_matching.ExtractKeypointsAndDescriptors(&image2.Data, imageAnalyzer)
	extractionTime := time1 + time2

	startTimeMatching := time.Now()
	matches := imageMatcher.FindMatches(imageDescriptors1, imageDescriptors2)

	imagesAreMatch, bestMatches := image_matching.DetermineSimilarity(matches, similarityThreshold)
	matchingTime := time.Since(startTimeMatching)

	if debug {
		image1Mat := image_matching.ConvertImageToMat(&image1.Data, color.RGBA{R: 255, G: 255, B: 255, A: 255})
		image2Mat := image_matching.ConvertImageToMat(&image2.Data, color.RGBA{R: 255, G: 255, B: 255, A: 255})
		file_handling.DrawMatches(&image1Mat, keypoints1, &image2Mat, keypoints2, *bestMatches)
	}

	return imagesAreMatch, keypoints1, keypoints2, extractionTime, matchingTime, nil
}

func getAnalyzerAndMatcher(analyzer, matcher string) (image_matching.FeatureImageAnalyzer, image_matching.ImageMatcher, error) {
	imageAnalyzer := image_matching.ImageAnalyzerMapping[analyzer]
	imageMatcher := image_matching.ImageMatcherMapping[matcher]

	if imageAnalyzer == nil || imageMatcher == nil {
		return nil, nil, errors.New("invalid option for analyzer or matcher")
	}
	return imageAnalyzer, imageMatcher, nil
}
