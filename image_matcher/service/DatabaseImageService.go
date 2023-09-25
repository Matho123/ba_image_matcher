package service

import (
	"errors"
	"fmt"
	"gocv.io/x/gocv"
	"image"
	"image-matcher/image_matcher/client"
	"image/color"
	"log"
	"time"
)

type RawImage struct {
	ExternalReference string
	Data              image.Image
}

var descriptorMapping = map[FeatureImageAnalyzer]string{
	SiftImageAnalyzer{}:  "sift_descriptor",
	ORBImageAnalyzer{}:   "orb_descriptor",
	BRISKImageAnalyzer{}: "brisk_descriptor",
}

const MaxChunkSize = 50

func AnalyzeAndSaveDatabaseImage(rawImages []*RawImage) error {
	databaseConnection, err := openDatabaseConnection()
	if err != nil {
		return err
	}
	defer databaseConnection.Close()

	for _, rawImage := range rawImages {

		_, siftDesc, _ := ExtractKeypointsAndDescriptors(&rawImage.Data, SiftImageAnalyzer{})
		_, orbDesc, _ := ExtractKeypointsAndDescriptors(&rawImage.Data, ORBImageAnalyzer{})
		_, briskDesc, _ := ExtractKeypointsAndDescriptors(&rawImage.Data, BRISKImageAnalyzer{})
		pHash, _ := client.GetPHashValue(rawImage.Data)

		err := insertImageIntoDatabaseSet(
			databaseConnection,
			DatabaseSetImage{
				externalReference: rawImage.ExternalReference,
				siftDescriptor:    convertImageMatToByteArray(siftDesc),
				orbDescriptor:     convertImageMatToByteArray(orbDesc),
				briskDescriptor:   convertImageMatToByteArray(briskDesc),
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
	searchImage RawImage,
	analyzer string,
	matcher string,
	similarityThreshold float64,
) ([]string, error, time.Duration, time.Duration) {
	imageAnalyzer, imageMatcher, err := getAnalyzerAndMatcher(analyzer, matcher)
	if err != nil {
		return []string{}, err, 0, 0
	}
	defer imageMatcher.close()

	databaseConnection, err := openDatabaseConnection()
	if err != nil {
		return []string{}, err, 0, 0
	}
	defer databaseConnection.Close()

	var matchingTime time.Duration

	_, searchImageDescriptor, extractionTime := ExtractKeypointsAndDescriptors(&searchImage.Data, imageAnalyzer)

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
			databaseImageDescriptor := convertByteArrayToMat(databaseImage.descriptor)
			matchingStart := time.Now()
			matches := imageMatcher.findMatches(searchImageDescriptor, databaseImageDescriptor)
			matchingTime += time.Since(matchingStart)

			isMatch, _ := determineSimilarity(matches, similarityThreshold)
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

func MatchImageAgainstDatabasePHash(searchImage RawImage, maxHammingDistance int) ([]string, error, time.Duration, time.Duration) {
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
			if hashesAreMatch(searchImageHash, databaseImage.hash, maxHammingDistance) {
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
	image1 RawImage,
	image2 RawImage,
	analyzer string,
	matcher string,
	similarityThreshold float64,
	debug bool,
) (bool, []gocv.KeyPoint, []gocv.KeyPoint, time.Duration, time.Duration, error) {
	var imagesAreMatch = false
	//var img image.Image
	//image_transformation.SaveImageToDisk("identical/test", image1.Data)
	//
	//img, _ = image_transformation.ResizeImage(&image1.Data)
	//image_transformation.SaveImageToDisk("scaled/test", img)
	//
	//img, _ = image_transformation.RotateImage(&image1.Data)
	//image_transformation.SaveImageToDisk("rotated/test", img)
	//
	//img, _ = image_transformation.MirrorImage(&image1.Data)
	//image_transformation.SaveImageToDisk("mirrored/test", img)
	//
	//img, _ = image_transformation.ChangeBackgroundColor(&image1.Data)
	//image_transformation.SaveImageToDisk("background/test", img)
	//
	//img, _ = image_transformation.MoveMotive(&image1.Data)
	//image_transformation.SaveImageToDisk("moved/test", img)
	//
	//img, _ = image_transformation.IntegrateInOtherImage(&image1.Data)
	//image_transformation.SaveImageToDisk("part/test", img)

	var extractionTime time.Duration
	var matchingTime time.Duration

	var keypoints1 []gocv.KeyPoint
	var keypoints2 []gocv.KeyPoint
	var time1 time.Duration
	var imageDescriptors1 gocv.Mat
	var imageDescriptors2 gocv.Mat
	var time2 time.Duration

	if analyzer == "phash" {
		hash1, extractionTime1 := client.GetPHashValue(image1.Data)
		hash2, extractionTime2 := client.GetPHashValue(image2.Data)
		extractionTime = time.Duration((extractionTime1 + extractionTime2) * float64(time.Second))

		log.Println(fmt.Sprintf("hash1: %d | hash2: %d", hash1, hash2))

		startTimeMatching := time.Now()
		imagesAreMatch = hashesAreMatch(hash1, hash2, 4)
		matchingTime = time.Since(startTimeMatching)
	} else {
		imageAnalyzer, imageMatcher, err := getAnalyzerAndMatcher(analyzer, matcher)
		if err != nil {
			return false, nil, nil, 0, 0, err
		}
		defer imageMatcher.close()

		keypoints1, imageDescriptors1, time1 = ExtractKeypointsAndDescriptors(&image1.Data, imageAnalyzer)
		keypoints2, imageDescriptors2, time2 = ExtractKeypointsAndDescriptors(&image2.Data, imageAnalyzer)
		extractionTime = time1 + time2

		startTimeMatching := time.Now()
		matches := imageMatcher.findMatches(imageDescriptors1, imageDescriptors2)

		var bestMatches []gocv.DMatch
		imagesAreMatch, bestMatches = determineSimilarity(matches, similarityThreshold)
		matchingTime = time.Since(startTimeMatching)

		if debug {
			image1Mat := convertImageToMat(&image1.Data, color.RGBA{R: 255, G: 255, B: 255, A: 255})
			image2Mat := convertImageToMat(&image2.Data, color.RGBA{R: 255, G: 255, B: 255, A: 255})
			drawMatches(&image1Mat, keypoints1, &image2Mat, keypoints2, bestMatches)
		}
	}

	return imagesAreMatch, keypoints1, keypoints2, extractionTime, matchingTime, nil
}

func getAnalyzerAndMatcher(analyzer, matcher string) (FeatureImageAnalyzer, ImageMatcher, error) {
	imageAnalyzer := imageAnalyzerMapping[analyzer]
	imageMatcher := imageMatcherMapping[matcher]

	if imageAnalyzer == nil || imageMatcher == nil {
		return nil, nil, errors.New("invalid option for analyzer or matcher")
	}
	return imageAnalyzer, imageMatcher, nil
}

func convertImageMatToByteArray(image gocv.Mat) []byte {
	if image.Empty() {
		log.Println("descriptor is empty!")
		return nil
	}

	nativeByteBuffer, err := gocv.IMEncode(".png", image)
	if err != nil {
		log.Println("unable to convert image to gocv.NativeByteBuffer! ", err)
		return nil
	}
	return nativeByteBuffer.GetBytes()
}

func convertByteArrayToMat(imageBytes []byte) gocv.Mat {
	imageMat, err := gocv.IMDecode(imageBytes, -1)

	if err != nil || imageMat.Empty() {
		log.Println("unable to convert bytes to gocv.mat")
	}
	return imageMat
}

func drawMatches(
	image1 *gocv.Mat,
	keypoints1 []gocv.KeyPoint,
	image2 *gocv.Mat,
	keypoints2 []gocv.KeyPoint,
	bestMatches []gocv.DMatch,
) {
	outImage := gocv.NewMat()
	gocv.DrawMatches(
		*image1,
		keypoints1,
		*image2,
		keypoints2,
		bestMatches,
		&outImage,
		color.RGBA{R: 255},
		color.RGBA{R: 255},
		[]byte{},
		gocv.DrawMatchesFlag(0),
	)
	gocv.IMWrite("debug/matches.png", outImage)
}