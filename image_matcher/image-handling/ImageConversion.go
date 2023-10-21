package image_handling

import (
	"gocv.io/x/gocv"
	"log"
)

func ConvertImageMatToByteArray(mat gocv.Mat) []byte {
	if mat.Empty() {
		log.Println("descriptor is empty!")
		return nil
	}
	return mat.ToBytes()
}

func ConvertImageDescriptorMat(descriptor *gocv.Mat, goalType gocv.MatType) *gocv.Mat {
	if descriptor.Type() != goalType {
		descriptor.ConvertTo(descriptor, goalType)
	}
	return descriptor
}

func ConvertByteArrayToDescriptorMat(descriptorBytes []byte, imageAnalyzer string) (*gocv.Mat, error) {
	switch imageAnalyzer {
	case SIFT:
		rows := len(descriptorBytes) / siftDescriptorByteLength
		return convertByteArrayToMat(descriptorBytes, rows, siftDescriptorByteLength/4, gocv.MatTypeCV32F)
	case ORB:
		rows := len(descriptorBytes) / orbDescriptorByteLength
		return convertByteArrayToMat(descriptorBytes, rows, orbDescriptorByteLength, gocv.MatTypeCV8U)
	case BRISK:
		rows := len(descriptorBytes) / briskDescriptorByteLength
		return convertByteArrayToMat(descriptorBytes, rows, briskDescriptorByteLength, gocv.MatTypeCV8U)
	default:
		return nil, nil
	}
}

func convertByteArrayToMat(bytes []byte, rows, cols int, matType gocv.MatType) (*gocv.Mat, error) {
	mat, err := gocv.NewMatFromBytes(rows, cols, matType, bytes)
	if err != nil || mat.Empty() {
		log.Println("unable to convert bytes to gocv.mat")
		return nil, err
	}
	return &mat, nil
}
