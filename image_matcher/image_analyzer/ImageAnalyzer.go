package image_analyzer

import (
	"gocv.io/x/gocv"
	"image"
	"image/color"
	"image_matcher/image_handling"
	"time"
)

const SIFT = "sift"
const ORB = "orb"
const BRISK = "brisk"
const PHASH = "phash"

var AnalyzerMapping = map[string]FeatureBasedImageAnalyzer{
	SIFT:  &SiftImageAnalyzer{gocv.NewSIFT()},
	ORB:   &ORBImageAnalyzer{gocv.NewORB()},
	BRISK: &BRISKImageAnalyzer{gocv.NewBRISK()},
}

type FeatureBasedImageAnalyzer interface {
	AnalyzeImage(image *gocv.Mat) ([]gocv.KeyPoint, gocv.Mat, time.Duration)
}

type SiftImageAnalyzer struct {
	analyzer gocv.SIFT
}

func (sift *SiftImageAnalyzer) AnalyzeImage(image *gocv.Mat) ([]gocv.KeyPoint, gocv.Mat, time.Duration) {
	startTime := time.Now()
	keypoints, descriptors := sift.analyzer.DetectAndCompute(*image, gocv.NewMat())
	extractionTime := time.Since(startTime)

	return keypoints, descriptors, extractionTime
}

type ORBImageAnalyzer struct {
	analyzer gocv.ORB
}

func (orb *ORBImageAnalyzer) AnalyzeImage(image *gocv.Mat) ([]gocv.KeyPoint, gocv.Mat, time.Duration) {
	startTime := time.Now()
	keypoints, descriptors := orb.analyzer.DetectAndCompute(*image, gocv.NewMat())
	extractionTime := time.Since(startTime)

	return keypoints, descriptors, extractionTime
}

type BRISKImageAnalyzer struct {
	analyzer gocv.BRISK
}

func (brisk *BRISKImageAnalyzer) AnalyzeImage(image *gocv.Mat) ([]gocv.KeyPoint, gocv.Mat, time.Duration) {
	startTime := time.Now()
	keypoints, descriptors := brisk.analyzer.DetectAndCompute(*image, gocv.NewMat())
	extractionTime := time.Since(startTime)

	return keypoints, descriptors, extractionTime
}

type PHash struct{}

func (PHash) GetHash(image image.Image) uint64 {
	return calculateHash(image)
}

func ExtractKeypointsAndDescriptors(img *image.Image, imageAnalyzer *FeatureBasedImageAnalyzer) (
	[]gocv.KeyPoint,
	gocv.Mat,
	time.Duration,
) {
	blackBgMat := image_handling.ConvertImageToMat(img, color.RGBA{A: 255})
	whiteBgMat := image_handling.ConvertImageToMat(img, color.RGBA{R: 255, G: 255, B: 255, A: 255})

	var finalKeypoints []gocv.KeyPoint
	var finalExtractionTime time.Duration
	var finalDescriptorByteArray gocv.Mat

	keypoints1, descriptorMat1, extractionTime1 := (*imageAnalyzer).AnalyzeImage(&blackBgMat)
	keypoints2, descriptorMat2, extractionTime2 := (*imageAnalyzer).AnalyzeImage(&whiteBgMat)

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
