package image_service

import (
	"database/sql"
	"errors"
	"fmt"
	"gocv.io/x/gocv"
	"image/color"
	"image_matcher/image_analyzer"
	"image_matcher/image_database"
	"image_matcher/image_handling"
	"image_matcher/image_matching"
	"log"
	"time"
)

var descriptorMapping = map[string]string{
	image_analyzer.SIFT:  "sift_descriptor",
	image_analyzer.ORB:   "orb_descriptor",
	image_analyzer.BRISK: "brisk_descriptor",
}

func AnalyzeAndSaveDatabaseImage(rawImages []*image_handling.RawImage) error {
	var err error

	err = image_database.ApplyDatabaseOperation(func(databaseConnection *sql.DB) {
		for _, rawImage := range rawImages {

			sift := image_analyzer.AnalyzerMapping[image_analyzer.SIFT]
			_, siftDesc, _ := image_analyzer.ExtractKeypointsAndDescriptors(&rawImage.Data, &sift)

			orb := image_analyzer.AnalyzerMapping[image_analyzer.ORB]
			_, orbDesc, _ := image_analyzer.ExtractKeypointsAndDescriptors(&rawImage.Data, &orb)

			brisk := image_analyzer.AnalyzerMapping[image_analyzer.BRISK]
			_, briskDesc, _ := image_analyzer.ExtractKeypointsAndDescriptors(&rawImage.Data, &brisk)

			pHash, _ := image_analyzer.GetPHashValue(&rawImage.Data)
			rotationInvariantHash, _ := image_analyzer.CalculateOrientedPHash(&rawImage.Data)

			err = image_database.InsertImageIntoDatabaseSet(
				databaseConnection,
				image_database.ForbiddenImageCreation{
					ExternalReference:     rawImage.ExternalReference,
					SiftDescriptor:        image_handling.ConvertImageMatToByteArray(siftDesc),
					OrbDescriptor:         image_handling.ConvertImageMatToByteArray(orbDesc),
					BriskDescriptor:       image_handling.ConvertImageMatToByteArray(briskDesc),
					PHash:                 pHash,
					RotationInvariantHash: rotationInvariantHash,
				},
			)
		}
	})

	return err
}

func MatchImageAgainstDatabaseHybrid(searchImage *image_handling.RawImage, debug bool) (
	*[]string,
	int,
	error,
	time.Duration,
	time.Duration,
) {
	regularHash, extractionTime1 := image_analyzer.GetPHashValue(&searchImage.Data)

	start := time.Now()
	mirroredX, _ := image_handling.MirrorImage(&searchImage.Data, true)
	mirroredY, _ := image_handling.MirrorImage(&searchImage.Data, false)
	totalExtractionTime := time.Since(start)

	hashes, _ := image_analyzer.CalculateOrientedHashes(&searchImage.Data)
	mirroredXHashes, extractionTime2 := image_analyzer.CalculateOrientedHashes(&mirroredX)
	mirroredYHashes, extractionTime3 := image_analyzer.CalculateOrientedHashes(&mirroredY)

	sift := image_analyzer.AnalyzerMapping[image_analyzer.SIFT]

	_, searchImageDescriptors, extractionTime4 := image_analyzer.ExtractKeypointsAndDescriptors(&mirroredX, &sift)

	hashes = append(hashes, mirroredXHashes...)
	hashes = append(hashes, mirroredYHashes...)

	matchedReferences, poolSize, matchingTime :=
		image_matching.HybridImageMatcher(
			hashes,
			regularHash,
			&searchImageDescriptors,
			debug,
		)
	totalExtractionTime =
		totalExtractionTime + time.Duration(extractionTime1)*time.Second + extractionTime2 + extractionTime3 + extractionTime4
	return matchedReferences, poolSize, nil, totalExtractionTime, matchingTime
}

func MatchImageAgainstDatabasePHash(searchImage *image_handling.RawImage, maxHammingDistance int, debug bool) (
	*[]string,
	error,
	time.Duration,
	time.Duration,
) {
	var totalMatchingTime time.Duration
	searchImageHash, extractionTime := image_analyzer.GetPHashValue(&searchImage.Data)
	var matchedImages []string

	err := image_database.ApplyChunkedPHashRetrievalOperation(func(databaseImage image_database.PHashImageEntity) {
		if debug {
			println("\nComparing to " + databaseImage.ExternalReference)
		}
		isMatch, _, matchingTime :=
			image_matching.HashesAreMatch(searchImageHash, databaseImage.Hash, maxHammingDistance, debug)
		totalMatchingTime += matchingTime

		if isMatch {
			matchedImages = append(matchedImages, databaseImage.ExternalReference)
		}
	})

	if err != nil {
		return nil, err, time.Duration(0), time.Duration(0)
	}

	return &matchedImages, nil, extractionTime, totalMatchingTime
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

	_, searchImageDescriptor, extractionTime := image_analyzer.ExtractKeypointsAndDescriptors(
		&searchImage.Data,
		imageAnalyzer,
	)

	var matchedImages []string
	var totalMatchingTime time.Duration

	err = image_database.ApplyChunkedFeatureBasedRetrievalOperation(func(databaseImage image_database.FeatureImageEntity) {
		if debug {
			println("\nComparing to " + databaseImage.ExternalReference)
		}
		databaseImageDescriptor, err :=
			image_handling.ConvertByteArrayToDescriptorMat(&databaseImage.Descriptors, analyzer)

		if databaseImageDescriptor == nil || err != nil {
			println("Descriptor was empty", databaseImage.ExternalReference)
			return
		}

		matchingStart := time.Now()
		matches := (*imageMatcher).FindMatches(&searchImageDescriptor, databaseImageDescriptor)
		totalMatchingTime += time.Since(matchingStart)
		databaseImageDescriptor.Close()

		isMatch, _, _ := image_matching.DetermineSimilarity(matches, similarityThreshold, debug)
		if isMatch {
			matchedImages = append(matchedImages, databaseImage.ExternalReference)
		}
	}, descriptorMapping[analyzer])

	if err != nil {
		return nil, err, nil, time.Duration(0), time.Duration(0)
	}

	return &matchedImages, nil, &searchImageDescriptor, extractionTime, totalMatchingTime
}

func MatchAgainstDatabaseFeatureBasedWithMultipleThresholds(
	searchImage *image_handling.RawImage, analyzer, matcher string, thresholds *[]float64,
) (*map[float64][]string, error, *gocv.Mat, time.Duration, time.Duration) {
	imageAnalyzer, imageMatcher, err := getAnalyzerAndMatcher(analyzer, matcher)
	if err != nil {
		log.Println(err)
	}

	_, searchImageDescriptor, extractionTime := image_analyzer.ExtractKeypointsAndDescriptors(
		&searchImage.Data,
		imageAnalyzer,
	)

	var totalMatchingTime time.Duration
	matchedImagesPerThreshold := make(map[float64][]string)

	for _, threshold := range *thresholds {
		matchedImagesPerThreshold[threshold] = []string{}
	}

	err = image_database.ApplyChunkedFeatureBasedRetrievalOperation(func(databaseImage image_database.FeatureImageEntity) {
		databaseImageDescriptor, err :=
			image_handling.ConvertByteArrayToDescriptorMat(&databaseImage.Descriptors, analyzer)

		if databaseImageDescriptor == nil || err != nil {
			println("Descriptor was empty", databaseImage.ExternalReference)
			return
		}

		matchingStart := time.Now()
		matches := (*imageMatcher).FindMatches(&searchImageDescriptor, databaseImageDescriptor)
		totalMatchingTime += time.Since(matchingStart)
		databaseImageDescriptor.Close()

		image_matching.FindDescriptorMatchesPerThreshold(
			matches, &matchedImagesPerThreshold, databaseImage.ExternalReference,
		)
	}, descriptorMapping[analyzer])
	if err != nil {
		return nil, err, nil, time.Duration(0), time.Duration(0)
	}

	return &matchedImagesPerThreshold, nil, &searchImageDescriptor, extractionTime, totalMatchingTime
}

func MatchImageAgainstDatabasePHashWithMultipleThresholds(searchImage *image_handling.RawImage, thresholds *[]int) (
	*map[int][]string,
	error,
	time.Duration,
	time.Duration,
) {

	var totalMatchingTime time.Duration
	searchImageHash, extractionTime := image_analyzer.GetPHashValue(&searchImage.Data)
	matchedImagesPerThreshold := make(map[int][]string)

	for _, threshold := range *thresholds {
		matchedImagesPerThreshold[threshold] = []string{}
	}

	err := image_database.ApplyChunkedPHashRetrievalOperation(func(databaseImage image_database.PHashImageEntity) {
		matchingTime := image_matching.FindHashMatchesPerThreshold(
			searchImageHash, databaseImage.Hash, &matchedImagesPerThreshold, databaseImage.ExternalReference,
		)
		totalMatchingTime += matchingTime
	})

	if err != nil {
		return nil, err, time.Duration(0), time.Duration(0)
	}

	return &matchedImagesPerThreshold, nil, extractionTime, totalMatchingTime
}

func AnalyzeAndMatchTwoImagesHash(
	image1 image_handling.RawImage,
	image2 image_handling.RawImage,
	analyzer string,
	threshold int,
) (bool, time.Duration, time.Duration) {
	if analyzer == image_analyzer.PHASH {
		hash1, extractionTime1 := image_analyzer.GetPHashValue(&image1.Data)
		hash2, extractionTime2 := image_analyzer.GetPHashValue(&image2.Data)
		extractionTime := extractionTime1 + extractionTime2

		imagesAreMatch, _, matchingTime := image_matching.HashesAreMatch(hash1, hash2, threshold, true)

		return imagesAreMatch, extractionTime, matchingTime
	}
	if analyzer == image_analyzer.NewAnalyzer {
		hash, extractionTime1 := image_analyzer.CalculateOrientedPHash(&image1.Data)
		hashes, extractionTime2 := image_analyzer.CalculateOrientedHashes(&image2.Data)
		match, matchedHash, _, matchingTime := image_matching.MatchOrientedHashes(hash, hashes, threshold)

		log.Println(fmt.Sprintf("hash1: %d | hash2: %d", hash, matchedHash))

		return match, extractionTime1 + extractionTime2, matchingTime
	} else {
		log.Fatal("Couldn't find analyzer!")
		return false, 0, 0
	}
}

func AnalyzeAndMatchTwoImagesFeatureBased(
	image1 image_handling.RawImage,
	image2 image_handling.RawImage,
	analyzer string,
	matcher string,
	threshold float64,
	debug bool,
) (bool, []gocv.KeyPoint, []gocv.KeyPoint, time.Duration, time.Duration, error) {
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

	imagesAreMatch, _, bestMatches := image_matching.DetermineSimilarity(matches, threshold, true)
	matchingTime := time.Since(startTimeMatching)

	if debug {
		image1Mat := image_handling.ConvertImageToGrayMatWithBackground(&image1.Data, color.RGBA{R: 255, G: 255, B: 255, A: 255})
		image2Mat := image_handling.ConvertImageToGrayMatWithBackground(&image2.Data, color.RGBA{R: 255, G: 255, B: 255, A: 255})
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
