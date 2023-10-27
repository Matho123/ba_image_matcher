package image_service

import (
	"errors"
	"fmt"
	"gocv.io/x/gocv"
	"image/color"
	"image_matcher/image_analyzer"
	"image_matcher/image_handling"
	"image_matcher/image_matching"
	"log"
	"time"
)

var DescriptorMapping = map[string]string{
	image_analyzer.SIFT:  "sift_descriptor",
	image_analyzer.ORB:   "orb_descriptor",
	image_analyzer.BRISK: "brisk_descriptor",
}

const MaxChunkSize = 50

func AnalyzeAndSaveDatabaseImage(rawImages []*image_handling.RawImage) error {
	databaseConnection, err := openDatabaseConnection()
	if err != nil {
		return err
	}
	defer databaseConnection.Close()

	for _, rawImage := range rawImages {

		sift := image_analyzer.AnalyzerMapping[image_analyzer.SIFT]
		_, siftDesc, _ := image_analyzer.ExtractKeypointsAndDescriptors(&rawImage.Data, &sift)

		orb := image_analyzer.AnalyzerMapping[image_analyzer.ORB]
		_, orbDesc, _ := image_analyzer.ExtractKeypointsAndDescriptors(&rawImage.Data, &orb)

		brisk := image_analyzer.AnalyzerMapping[image_analyzer.BRISK]
		_, briskDesc, _ := image_analyzer.ExtractKeypointsAndDescriptors(&rawImage.Data, &brisk)
		pHash, _ := image_analyzer.GetPHashValue(rawImage.Data)

		err := insertImageIntoDatabaseSet(
			databaseConnection,
			ForbiddenImageCreation{
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
	searchImage *image_handling.RawImage,
	analyzer string,
	matcher string,
	similarityThreshold float64,
	debug bool,
) (*[]string, error, *gocv.Mat, time.Duration, time.Duration) {
	imageAnalyzer, imageMatcher, err := getAnalyzerAndMatcher(analyzer, matcher)
	if err != nil {
		return nil, err, nil, 0, 0
	}

	databaseConnection, err := openDatabaseConnection()
	if err != nil {
		return nil, err, nil, 0, 0
	}
	defer databaseConnection.Close()

	var totalMatchingTime time.Duration

	_, searchImageDescriptor, extractionTime := image_analyzer.ExtractKeypointsAndDescriptors(
		&searchImage.Data,
		imageAnalyzer,
	)

	var matchedImages []string

	offset := 0
	for {
		databaseImageChunk, err := retrieveFeatureImageChunk(
			databaseConnection,
			DescriptorMapping[analyzer],
			offset,
			MaxChunkSize+1,
		)
		if err != nil {
			log.Println("Error while retrieving chunk from database images: ", err)
		}

		for _, databaseImage := range (*databaseImageChunk)[0 : len(*databaseImageChunk)-1] {
			if debug {
				println("\nComparing to " + databaseImage.externalReference)
			}
			databaseImageDescriptor, err :=
				image_handling.ConvertByteArrayToDescriptorMat(&databaseImage.descriptors, analyzer)

			if databaseImageDescriptor == nil || err != nil {
				println("Descriptor was empty", databaseImage.externalReference)
				continue
			}

			matchingStart := time.Now()
			matches := (*imageMatcher).FindMatches(&searchImageDescriptor, databaseImageDescriptor)
			totalMatchingTime += time.Since(matchingStart)
			databaseImageDescriptor.Close()

			isMatch, _, _ := image_matching.DetermineSimilarity(matches, similarityThreshold, debug)
			if isMatch {
				matchedImages = append(matchedImages, databaseImage.externalReference)
			}
		}

		if len(*databaseImageChunk) < MaxChunkSize+1 {
			break
		}
		offset += MaxChunkSize

		databaseImageChunk = nil
	}
	return &matchedImages, nil, &searchImageDescriptor, extractionTime, totalMatchingTime
}

func MatchImageAgainstDatabasePHash(searchImage *image_handling.RawImage, maxHammingDistance int, debug bool) (
	*[]string,
	error,
	time.Duration,
	time.Duration,
) {
	databaseConnection, err := openDatabaseConnection()
	if err != nil {
		return nil, err, 0, 0
	}
	defer databaseConnection.Close()

	var totalMatchingTime time.Duration

	searchImageHash, extractionTime := image_analyzer.GetPHashValue(searchImage.Data)

	var matchedImages []string

	offset := 0
	for {
		databaseImageChunk, err := retrievePHashImageChunk(databaseConnection, offset, MaxChunkSize+1)
		if err != nil {
			log.Println("Error while retrieving chunk from database images: ", err)
		}

		for _, databaseImage := range (*databaseImageChunk)[0 : len(*databaseImageChunk)-1] {
			if debug {
				println("\nComparing to " + databaseImage.externalReference)
			}

			isMatch, matchingTime :=
				image_matching.HashesAreMatch(searchImageHash, databaseImage.hash, maxHammingDistance, debug)
			totalMatchingTime += matchingTime

			if isMatch {
				matchedImages = append(matchedImages, databaseImage.externalReference)
			}
		}

		if len(*databaseImageChunk) < MaxChunkSize+1 {
			break
		}
		offset += MaxChunkSize
	}
	return &matchedImages, nil, time.Duration(extractionTime * float64(time.Second)), totalMatchingTime
}

func AnalyzeAndMatchTwoImages(
	image1 image_handling.RawImage,
	image2 image_handling.RawImage,
	analyzer string,
	matcher string,
	similarityThreshold float64,
	debug bool,
) (bool, []gocv.KeyPoint, []gocv.KeyPoint, time.Duration, time.Duration, error) {
	if analyzer == image_analyzer.PHASH {
		hash1, extractionTime1 := image_analyzer.GetPHashValue(image1.Data)
		hash2, extractionTime2 := image_analyzer.GetPHashValue(image2.Data)
		extractionTime := time.Duration((extractionTime1 + extractionTime2) * float64(time.Second))

		log.Println(fmt.Sprintf("hash1: %d | hash2: %d", hash1, hash2))
		//hash12 := image_handling.PHash{}.GetHash(image1.Data)
		//hash22 := image_handling.PHash{}.GetHash(image2.Data)
		//log.Println(
		//	fmt.Sprintf(
		//		"hash1: %d | hash2: %d",
		//		hash12,
		//		hash22,
		//	),
		//)

		imagesAreMatch, matchingTime := image_matching.HashesAreMatch(hash1, hash2, 4, true)

		return imagesAreMatch, []gocv.KeyPoint{}, []gocv.KeyPoint{}, extractionTime, matchingTime, nil
	}

	//for feature-based analyzer
	imageAnalyzer, imageMatcher, err := getAnalyzerAndMatcher(analyzer, matcher)
	if err != nil {
		return false, nil, nil, 0, 0, err
	}

	keypoints1, imageDescriptors1, time1 := image_analyzer.ExtractKeypointsAndDescriptors(&image1.Data, imageAnalyzer)
	defer imageDescriptors1.Close()
	keypoints2, imageDescriptors2, time2 := image_analyzer.ExtractKeypointsAndDescriptors(&image2.Data, imageAnalyzer)
	defer imageDescriptors2.Close()
	extractionTime := time1 + time2

	startTimeMatching := time.Now()
	matches := (*imageMatcher).FindMatches(&imageDescriptors1, &imageDescriptors2)

	imagesAreMatch, _, bestMatches := image_matching.DetermineSimilarity(matches, similarityThreshold, true)
	matchingTime := time.Since(startTimeMatching)

	if debug {
		image1Mat := image_handling.ConvertImageToMat(&image1.Data, color.RGBA{R: 255, G: 255, B: 255, A: 255})
		image2Mat := image_handling.ConvertImageToMat(&image2.Data, color.RGBA{R: 255, G: 255, B: 255, A: 255})
		image_handling.DrawMatches(&image1Mat, keypoints1, &image2Mat, keypoints2, bestMatches)
	}

	return imagesAreMatch, keypoints1, keypoints2, extractionTime, matchingTime, nil
}

func getAnalyzerAndMatcher(analyzer, matcher string) (*image_analyzer.FeatureBasedImageAnalyzer, *image_matching.FeatureBasedImageMatcher, error) {
	imageAnalyzer := image_analyzer.AnalyzerMapping[analyzer]
	imageMatcher := image_matching.MatcherMapping[matcher]

	if imageAnalyzer == nil || imageMatcher == nil {
		return nil, nil, errors.New("couldn't find analyzer or matcher")
	}

	return &imageAnalyzer, &imageMatcher, nil
}
