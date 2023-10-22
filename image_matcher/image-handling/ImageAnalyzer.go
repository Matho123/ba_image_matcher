package image_handling

import (
	"errors"
	"gocv.io/x/gocv"
	"image"
	"image/color"
	"time"
)

const SIFT = "sift"
const ORB = "orb"
const BRISK = "brisk"
const PHASH = "phash"

func GetFeatureBasedAnalyzer(analyzer string) (*FeatureBasedImageAnalyzer, error) {
	switch analyzer {
	case SIFT:
		return NewSiftAnalyzer(), nil
	case ORB:
		return NewOrbAnalyzer(), nil
	case BRISK:
		return NewBriskAnalyzer(), nil
	default:
		return nil, errors.New("invalid option for analyzer")
	}
}

type FeatureBasedImageAnalyzer interface {
	AnalyzeImage(image *gocv.Mat) ([]gocv.KeyPoint, gocv.Mat, time.Duration)
	Close()
}

type SiftImageAnalyzer struct {
	analyzer gocv.SIFT
}

func NewSiftAnalyzer() *FeatureBasedImageAnalyzer {
	var analyzer FeatureBasedImageAnalyzer
	analyzer = &SiftImageAnalyzer{gocv.NewSIFT()}
	return &analyzer
}

func (sift *SiftImageAnalyzer) AnalyzeImage(image *gocv.Mat) ([]gocv.KeyPoint, gocv.Mat, time.Duration) {
	startTime := time.Now()
	keypoints, descriptors := sift.analyzer.DetectAndCompute(*image, gocv.NewMat())
	extractionTime := time.Since(startTime)

	return keypoints, descriptors, extractionTime
}

func (sift *SiftImageAnalyzer) Close() {
	sift.analyzer.Close()
}

type ORBImageAnalyzer struct {
	analyzer gocv.ORB
}

func NewOrbAnalyzer() *FeatureBasedImageAnalyzer {
	var analyzer FeatureBasedImageAnalyzer
	analyzer = &ORBImageAnalyzer{gocv.NewORB()}
	return &analyzer
}

func (orb *ORBImageAnalyzer) AnalyzeImage(image *gocv.Mat) ([]gocv.KeyPoint, gocv.Mat, time.Duration) {
	startTime := time.Now()
	keypoints, descriptors := orb.analyzer.DetectAndCompute(*image, gocv.NewMat())
	extractionTime := time.Since(startTime)

	return keypoints, descriptors, extractionTime
}

func (orb *ORBImageAnalyzer) Close() {
	orb.analyzer.Close()
}

type BRISKImageAnalyzer struct {
	analyzer gocv.BRISK
}

func NewBriskAnalyzer() *FeatureBasedImageAnalyzer {
	var analyzer FeatureBasedImageAnalyzer
	analyzer = &BRISKImageAnalyzer{gocv.NewBRISK()}
	return &analyzer
}

func (brisk *BRISKImageAnalyzer) AnalyzeImage(image *gocv.Mat) ([]gocv.KeyPoint, gocv.Mat, time.Duration) {
	startTime := time.Now()
	keypoints, descriptors := brisk.analyzer.DetectAndCompute(*image, gocv.NewMat())
	extractionTime := time.Since(startTime)

	return keypoints, descriptors, extractionTime
}

func (brisk *BRISKImageAnalyzer) Close() {
	brisk.analyzer.Close()
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
	blackBgMat := ConvertImageToMat(img, color.RGBA{A: 255})
	whiteBgMat := ConvertImageToMat(img, color.RGBA{R: 255, G: 255, B: 255, A: 255})

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
