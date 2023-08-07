package service

import "gocv.io/x/gocv"

var imageAnalyzerMapping = map[string]ImageAnalyzer{
	"orb":  ORBImageAnalyzer{},
	"sift": SiftImageAnalyzer{},
}

type ImageAnalyzer interface {
	analyzeImage(image gocv.Mat) ([]gocv.KeyPoint, gocv.Mat)
}

type SiftImageAnalyzer struct{}

func (SIFT SiftImageAnalyzer) analyzeImage(image gocv.Mat) ([]gocv.KeyPoint, gocv.Mat) {
	sift := gocv.NewSIFT()
	defer sift.Close()

	//TODO: different Image for mask and original image in color
	return sift.DetectAndCompute(image, image)
}

type ORBImageAnalyzer struct{}

func (ORB ORBImageAnalyzer) analyzeImage(image gocv.Mat) ([]gocv.KeyPoint, gocv.Mat) {
	orb := gocv.NewORB()
	defer orb.Close()

	return orb.DetectAndCompute(image, image)
}

func extractKeypointsAndDescriptors(imageMatPointer gocv.Mat, imageAnalyzer ImageAnalyzer) ([]gocv.KeyPoint, []byte) {
	keypoints, descriptorMatPointer := imageAnalyzer.analyzeImage(imageMatPointer)
	gocv.IMWrite("test/images/test.png", descriptorMatPointer)
	defer imageMatPointer.Close()

	descriptorByteArray := convertImageMatToByteArray(descriptorMatPointer)
	defer descriptorMatPointer.Close()

	return keypoints, descriptorByteArray
}
