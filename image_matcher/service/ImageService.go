package service

import (
	"errors"
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

func AnalyzeAndSave(rawImages []*RawImage) error {
	databaseConnection, err := openDatabaseConnection()
	if err != nil {
		return err
	}
	defer databaseConnection.Close()

	for _, rawImage := range rawImages {
		imageMat := convertImageToMat(&rawImage.Data)

		_, siftDesc := extractKeypointsAndDescriptors(&imageMat, SiftImageAnalyzer{})
		_, orbDesc := extractKeypointsAndDescriptors(&imageMat, ORBImageAnalyzer{})
		_, briskDesc := extractKeypointsAndDescriptors(&imageMat, BRISKImageAnalyzer{})
		pHash := client.GetPHashValue(rawImage.Data)

		err := insertImageIntoDatabaseSet(
			databaseConnection,
			DatabaseSetImage{
				externalReference: rawImage.ExternalReference,
				siftDescriptor:    siftDesc,
				orbDescriptor:     orbDesc,
				briskDescriptor:   briskDesc,
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

	databaseConnection, err := openDatabaseConnection()
	if err != nil {
		return []string{}, err, 0, 0
	}
	defer databaseConnection.Close()

	var extractionTime time.Duration
	var matchingTime time.Duration

	imageMat := convertImageToMat(&searchImage.Data)

	extractionStart := time.Now()
	_, searchImageDescriptor := imageAnalyzer.analyzeImage(&imageMat)
	extractionTime = time.Since(extractionStart)

	var matchedImages []string

	offset := 0
	for {
		databaseImages, err := retrieveFeatureImageChunk(
			databaseConnection,
			descriptorMapping[imageAnalyzer],
			offset,
			MaxChunkSize+1)
		if err != nil {
			log.Println("Error while retrieving chunk from database images: ", err)
		}

		matchingStart := time.Now()
		for _, databaseImage := range databaseImages {
			databaseImageDescriptor := convertByteArrayToMat(databaseImage.descriptor)
			matches := imageMatcher.findMatches(searchImageDescriptor, databaseImageDescriptor)

			isMatch, _ := determineSimilarity(matches, similarityThreshold)
			if isMatch {
				matchedImages = append(matchedImages, databaseImage.externalReference)
			}
		}
		matchingTime += time.Since(matchingStart)

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

	var extractionTime time.Duration
	var matchingTime time.Duration

	extractionStart := time.Now()
	searchImageHash := client.GetPHashValue(searchImage.Data)
	extractionTime = time.Since(extractionStart)

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
	return matchedImages, nil, extractionTime, matchingTime
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
	//image_transformation.ResizeImage(&image1.Data, image1.Data.Bounds().Dx()/2, image1.Data.Bounds().Dy()/2)
	//image_transformation.RotateImage(&image1.Data, 45.0)
	//image_transformation.MirrorImage(&image1.Data, true)
	//image_transformation.ChangeBackgroundColor(&image1.Data, color.RGBA{R: 255, A: 255})
	var extractionTime time.Duration
	var matchingTime time.Duration

	var keypoints1 []gocv.KeyPoint
	var keypoints2 []gocv.KeyPoint
	var imageDescriptors1 gocv.Mat
	var imageDescriptors2 gocv.Mat

	if analyzer == "phash" {
		startTimeExtraction := time.Now()
		hash1 := client.GetPHashValue(image1.Data)
		hash2 := client.GetPHashValue(image2.Data)
		extractionTime = time.Since(startTimeExtraction)

		startTimeMatching := time.Now()
		imagesAreMatch = hashesAreMatch(hash1, hash2, 4)
		matchingTime = time.Since(startTimeMatching)
	} else {
		imageAnalyzer, imageMatcher, err := getAnalyzerAndMatcher(analyzer, matcher)
		if err != nil {
			return false, nil, nil, 0, 0, err
		}

		image1Mat := convertImageToMat(&image1.Data)
		image2Mat := convertImageToMat(&image2.Data)

		startTimeExtraction := time.Now()
		keypoints1, imageDescriptors1 = imageAnalyzer.analyzeImage(&image1Mat)
		keypoints2, imageDescriptors2 = imageAnalyzer.analyzeImage(&image2Mat)
		extractionTime = time.Since(startTimeExtraction)

		startTimeMatching := time.Now()
		matches := imageMatcher.findMatches(imageDescriptors1, imageDescriptors2)

		var bestMatches []gocv.DMatch
		imagesAreMatch, bestMatches = determineSimilarity(matches, similarityThreshold)
		matchingTime = time.Since(startTimeMatching)

		if debug {
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
	nativeByteBuffer, err := gocv.IMEncode(".png", image)

	if err != nil {
		log.Println("unable to convert image to gocv.NativeByteBuffer! ", err)
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

func convertImageToMat(image *image.Image) gocv.Mat {
	mat, err := gocv.ImageToMatRGBA(*image)
	if err != nil {
		log.Fatalf("Error converting image to Mat: %v", err)
	}
	gocv.CvtColor(mat, &mat, gocv.ColorRGBAToGray)
	return mat
}
