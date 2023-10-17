package image_matching

import (
	"github.com/disintegration/imaging"
	"gocv.io/x/gocv"
	"image"
	"image/color"
	"log"
	"time"
)

const SIFT = "sift"
const ORB = "orb"
const BRISK = "brisk"
const PHASH = "phash"

var ImageAnalyzerMapping = map[string]FeatureImageAnalyzer{
	SIFT:  SiftImageAnalyzer{},
	ORB:   ORBImageAnalyzer{},
	BRISK: BRISKImageAnalyzer{},
}

type FeatureImageAnalyzer interface {
	AnalyzeImage(image *gocv.Mat) ([]gocv.KeyPoint, gocv.Mat)
}

type SiftImageAnalyzer struct{}

func (SiftImageAnalyzer) AnalyzeImage(image *gocv.Mat) ([]gocv.KeyPoint, gocv.Mat) {
	sift := gocv.NewSIFT()
	defer sift.Close()

	return sift.DetectAndCompute(*image, gocv.NewMat())
}

type ORBImageAnalyzer struct{}

func (ORBImageAnalyzer) AnalyzeImage(image *gocv.Mat) ([]gocv.KeyPoint, gocv.Mat) {
	orb := gocv.NewORB()
	defer orb.Close()

	return orb.DetectAndCompute(*image, gocv.NewMat())
}

type BRISKImageAnalyzer struct{}

func (BRISKImageAnalyzer) AnalyzeImage(image *gocv.Mat) ([]gocv.KeyPoint, gocv.Mat) {
	brisk := gocv.NewBRISK()
	defer brisk.Close()

	return brisk.DetectAndCompute(*image, gocv.NewMat())
}

type PHash struct{}

func (PHash) GetHash(image image.Image) uint64 {
	return calculateHash(image)
}

func ExtractKeypointsAndDescriptors(img *image.Image, imageAnalyzer FeatureImageAnalyzer) (
	[]gocv.KeyPoint,
	gocv.Mat,
	time.Duration,
) {
	blackBgMat := ConvertImageToMat(img, color.RGBA{A: 255})
	whiteBgMat := ConvertImageToMat(img, color.RGBA{R: 255, G: 255, B: 255, A: 255})

	var finalKeypoints []gocv.KeyPoint
	var finalExtractionTime time.Duration
	var finalDescriptorByteArray gocv.Mat

	startTime1 := time.Now()
	keypoints1, descriptorMat1 := imageAnalyzer.AnalyzeImage(&blackBgMat)
	extractionTime1 := time.Since(startTime1)

	startTime2 := time.Now()
	keypoints2, descriptorMat2 := imageAnalyzer.AnalyzeImage(&whiteBgMat)
	extractionTime2 := time.Since(startTime2)

	if len(keypoints1) > len(keypoints2) {
		finalKeypoints = keypoints1
		finalDescriptorByteArray = descriptorMat1
		finalExtractionTime = extractionTime1
	} else {
		finalKeypoints = keypoints2
		finalDescriptorByteArray = descriptorMat2
		finalExtractionTime = extractionTime2
	}
	return finalKeypoints, finalDescriptorByteArray, finalExtractionTime
}

func ConvertImageToMat(img *image.Image, c color.Color) gocv.Mat {
	newImage := imaging.New((*img).Bounds().Size().X, (*img).Bounds().Size().Y, c)
	newImage = imaging.Overlay(newImage, *img, image.Pt(0, 0), 1.0)

	mat1, err := gocv.ImageToMatRGBA(newImage)

	if err != nil {
		log.Println("Error converting image to Mat: ", err)
	}

	gocv.CvtColor(mat1, &mat1, gocv.ColorRGBAToGray)

	return mat1
}

//func convertImageToGrayWithAlpha(img *gocv.Mat) gocv.Mat {
//	// Create a new Mat for the grayscale image with alpha channel preserved
//	grayWithAlpha := gocv.NewMatWithSize(img.Rows(), img.Cols(), gocv.MatTypeCV8UC4)
//
//	// Split the RGBA image into separate color and alpha channels
//	channels := gocv.Split(*img)
//
//	// Create a blank single-channel image for the grayscale data
//	gray := gocv.NewMatWithSize(img.Rows(), img.Cols(), gocv.MatTypeCV8UC1)
//
//	// Convert the original image to grayscale (you can choose a different method)
//	gocv.CvtColor(*img, &gray, gocv.ColorRGBAToGray)
//
//	// Merge the grayscale data with the alpha channel into a four-channel image
//	gocv.Merge([]gocv.Mat{gray, gray, gray, channels[3]}, &grayWithAlpha)
//
//	// Release resources
//	for _, ch := range channels {
//		ch.Close()
//	}
//	gray.Close()
//
//	return grayWithAlpha
//}
