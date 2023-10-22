package image_handling

import (
	"github.com/disintegration/imaging"
	"gocv.io/x/gocv"
	"image"
	"image/color"
	"log"
)

// sift descriptors consist of 128 32-bit floating point numbers
// a 32-bit values be represented 4 bytes (32 / 8 = 4)
// so a sift descriptor needs 128 * 4 bytes
const siftDescriptorByteLength = 128 * 4

// orb and brisk descriptors are binary strings
// orb descriptors have 256 bit and brisk descriptors have 512 bit
// when converting them to bytes their length is divided by 8
const orbDescriptorByteLength = 256 / 8
const briskDescriptorByteLength = 512 / 8

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

func ConvertByteArrayToDescriptorMat(descriptorBytes *[]byte, imageAnalyzer string) (*gocv.Mat, error) {
	switch imageAnalyzer {
	case SIFT:
		rows := len(*descriptorBytes) / siftDescriptorByteLength
		return convertByteArrayToMat(*descriptorBytes, rows, siftDescriptorByteLength/4, gocv.MatTypeCV32F)
	case ORB:
		rows := len(*descriptorBytes) / orbDescriptorByteLength
		return convertByteArrayToMat(*descriptorBytes, rows, orbDescriptorByteLength, gocv.MatTypeCV8U)
	case BRISK:
		rows := len(*descriptorBytes) / briskDescriptorByteLength
		return convertByteArrayToMat(*descriptorBytes, rows, briskDescriptorByteLength, gocv.MatTypeCV8U)
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
