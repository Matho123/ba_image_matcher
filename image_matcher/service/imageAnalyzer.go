package service

import "gocv.io/x/gocv"

var imageAnalyzerMapping = map[string]ImageAnalyzer{
	"orb":   ORBImageAnalyzer{},
	"sift":  SiftImageAnalyzer{},
	"akaze": AKAZEImageAnalyzer{},
}

type ImageAnalyzer interface {
	analyzeImage(image *gocv.Mat) ([]gocv.KeyPoint, gocv.Mat)
}

type SiftImageAnalyzer struct{}

func (SiftImageAnalyzer) analyzeImage(image *gocv.Mat) ([]gocv.KeyPoint, gocv.Mat) {
	sift := gocv.NewSIFT()
	defer sift.Close()

	return sift.DetectAndCompute(*image, gocv.NewMat())
}

type ORBImageAnalyzer struct{}

func (ORBImageAnalyzer) analyzeImage(image *gocv.Mat) ([]gocv.KeyPoint, gocv.Mat) {
	orb := gocv.NewORB()
	defer orb.Close()

	return orb.DetectAndCompute(*image, gocv.NewMat())
}

type AKAZEImageAnalyzer struct{}

func (AKAZEImageAnalyzer) analyzeImage(image *gocv.Mat) ([]gocv.KeyPoint, gocv.Mat) {
	akaze := gocv.NewAKAZE()
	defer akaze.Close()

	return akaze.DetectAndCompute(*image, gocv.NewMat())
}

func extractKeypointsAndDescriptors(imageMatPointer *gocv.Mat, imageAnalyzer ImageAnalyzer) ([]gocv.KeyPoint, []byte) {
	keypoints, descriptorMatPointer := imageAnalyzer.analyzeImage(imageMatPointer)
	defer imageMatPointer.Close()

	descriptorByteArray := convertImageMatToByteArray(descriptorMatPointer)
	defer descriptorMatPointer.Close()

	return keypoints, descriptorByteArray
}
