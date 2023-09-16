package service

import (
	"errors"
	"gocv.io/x/gocv"
	"image"
	"image-matcher/image_matcher/client"
	"image/color"
	"log"
)

type RawImage struct {
	ExternalReference string
	Data              image.Image
}

type ProcessedImage struct {
	externalReference string
	descriptorData    []byte
}

const MaxChunkSize = 50

func AnalyzeAndSave(rawImages []*RawImage, analyzer string) error {
	imageAnalyzer := imageAnalyzerMapping[analyzer]
	if imageAnalyzer == nil {
		return errors.New("invalid image analyzer")
	}

	databaseConnection, err := openDatabaseConnection()

	if err != nil {
		return err
	}
	defer databaseConnection.Close()

	for _, rawImage := range rawImages {
		imageMat := convertImageToMat(&rawImage.Data)
		_, descriptor := extractKeypointsAndDescriptors(&imageMat, imageAnalyzer)

		processedImage := ProcessedImage{externalReference: rawImage.ExternalReference, descriptorData: descriptor}

		err := insertImageIntoDatabase(
			databaseConnection,
			processedImage,
		)

		if err != nil {
			return err
		}
	}
	return nil
}

// MatchAgainstDatabase TODO: add option to also register Image in the database if no match is found
func MatchAgainstDatabase(rawQueryImage RawImage, analyzer string, matcher string) (string, error) {
	imageAnalyzer, imageMatcher, err := getAnalyzerAndMatcher(analyzer, matcher)
	if err != nil {
		return "", err
	}

	databaseConnection, err := openDatabaseConnection()
	if err != nil {
		return "", err
	}
	defer databaseConnection.Close()

	imageMat := convertImageToMat(&rawQueryImage.Data)

	offset := 0
	for {
		databaseImages, err := retrieveImageChunkFromDatabase(databaseConnection, offset, MaxChunkSize+1)
		if err != nil {
			return "", err
		}

		_, queryImageDescriptor := imageAnalyzer.analyzeImage(&imageMat)
		for _, databaseImage := range databaseImages {
			databaseImageDescriptor := convertByteArrayToMat(databaseImage.descriptorData)
			matches := imageMatcher.findMatches(queryImageDescriptor, databaseImageDescriptor)

			isMatch, _ := determineSimilarity(matches)
			if isMatch {
				return databaseImage.externalReference, nil
			}
		}

		if len(databaseImages) < MaxChunkSize+1 {
			break
		}
		offset += MaxChunkSize
	}
	return "", nil
}

func AnalyzeAndMatchTwoImages(
	image1 RawImage,
	image2 RawImage,
	analyzer string,
	matcher string,
	debug bool,
) (bool, error) {
	var imagesAreMatch bool = false
	if analyzer == "phash" {
		hash1 := client.GetPHashValue(image1.Data)
		hash2 := client.GetPHashValue(image2.Data)
		imagesAreMatch = hashesAreMatch(hash1, hash2)
	} else {
		imageAnalyzer, imageMatcher, err := getAnalyzerAndMatcher(analyzer, matcher)
		if err != nil {
			return false, err
		}

		image1Mat := convertImageToMat(&image1.Data)
		image2Mat := convertImageToMat(&image2.Data)

		keypoints1, imageDescriptors1 := imageAnalyzer.analyzeImage(&image1Mat)
		keypoints2, imageDescriptors2 := imageAnalyzer.analyzeImage(&image2Mat)

		matches := imageMatcher.findMatches(imageDescriptors1, imageDescriptors2)

		var bestMatches []gocv.DMatch
		imagesAreMatch, bestMatches = determineSimilarity(matches)

		if debug {
			drawMatches(&image1Mat, keypoints1, &image2Mat, keypoints2, bestMatches)
		}
	}

	return imagesAreMatch, nil
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
