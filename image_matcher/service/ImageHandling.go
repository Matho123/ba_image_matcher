package service

import (
	"gocv.io/x/gocv"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	image_matching "image_matcher/image-matching"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
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

type RawImage struct {
	ExternalReference string
	Data              image.Image
}

var allowedImageExtensions = [...]string{".png", ".jpg"}

func LoadImagesFromPath(path string) []*RawImage {
	fileInfo, err := os.Stat(path)

	if err != nil {
		log.Println(err)
		return []*RawImage{}
	}

	if fileInfo.IsDir() {
		paths := GetFilePathsFromDirectory(path)
		return LoadImagesFromDirectory(paths)
	} else {
		return []*RawImage{LoadRawImage(path)}
	}

}

func GetFilePathsFromDirectory(directoryPath string) []string {
	var filePaths []string

	err := filepath.Walk(directoryPath, func(filePath string, fileInfo fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !fileInfo.IsDir() {
			filePaths = append(filePaths, filePath)
		}

		return nil
	})

	if err != nil {
		log.Println(err)
	}
	return filePaths
}

func LoadImagesFromDirectory(filePaths []string) []*RawImage {
	var rawImageDtos []*RawImage

	for _, path := range filePaths {
		rawImageDto := LoadRawImage(path)
		if rawImageDto != nil {
			rawImageDtos = append(rawImageDtos, rawImageDto)
		}
	}

	return rawImageDtos
}

func LoadRawImage(path string) *RawImage {
	if !isAllowedImageFile(path) {
		return nil
	}

	img := LoadImageFromDisk(path)

	filenameWithExt := filepath.Base(path)
	filenameWithoutExt := strings.TrimSuffix(filenameWithExt, filepath.Ext(filenameWithExt))

	return &RawImage{ExternalReference: filenameWithoutExt, Data: *img}
}

func LoadImageFromDisk(path string) *image.Image {
	//image := gocv.IMRead(filePath, gocv.IMReadGrayScale)
	file, err := os.Open(path)
	if err != nil {
		log.Fatal("Error opening the image: ", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		log.Fatal("Error decoding the image: ", err)
	}

	//log.Println("successfully loaded: ", path)
	return &img
}

func SaveImageToDisk(name string, image image.Image) {
	newPath := "images/variations/" + name + ".png"
	outputFile, err := os.Create(newPath)
	if err != nil {
		log.Println("Error while creating outputfile for image: ", err)
		return
	}
	defer outputFile.Close()

	err = png.Encode(outputFile, image)
	if err != nil {
		log.Println("Error while saving image "+name+" to disk: ", err)
		return
	}
	log.Println("saved variation", newPath)
}

func isAllowedImageFile(filePath string) bool {
	fileExtension := strings.ToLower(filepath.Ext(filePath))

	for _, allowedExtension := range allowedImageExtensions {
		if allowedExtension == fileExtension {
			return true
		}
	}
	return false
}

func ConvertImageMatToByteArray(mat gocv.Mat) []byte {
	if mat.Empty() {
		log.Println("descriptor is empty!")
		return nil
	}
	return mat.ToBytes()
}

func ConvertByteArrayToDescriptorMat(descriptorBytes []byte, imageAnalyzer string) *gocv.Mat {
	switch imageAnalyzer {
	case image_matching.SIFT:
		rows := len(descriptorBytes) / siftDescriptorByteLength
		return convertByteArrayToMat(descriptorBytes, rows, siftDescriptorByteLength/4, gocv.MatTypeCV32F)
	case image_matching.ORB:
		rows := len(descriptorBytes) / orbDescriptorByteLength
		return convertByteArrayToMat(descriptorBytes, rows, orbDescriptorByteLength, gocv.MatTypeCV8U)
	case image_matching.BRISK:
		rows := len(descriptorBytes) / briskDescriptorByteLength
		return convertByteArrayToMat(descriptorBytes, rows, briskDescriptorByteLength, gocv.MatTypeCV8U)
	default:
		return nil
	}
}

func convertByteArrayToMat(bytes []byte, rows, cols int, matType gocv.MatType) *gocv.Mat {
	log.Println(bytes)
	mat, err := gocv.NewMatFromBytes(rows, cols, matType, bytes)
	if err != nil || mat.Empty() {
		log.Println(mat.Empty(), err)
		log.Println("unable to convert bytes to gocv.mat")
	}
	return &mat
}

func DrawMatches(
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
		color.RGBA{R: 255, A: 100},
		color.RGBA{R: 255},
		[]byte{},
		gocv.DrawMatchesFlag(0),
	)
	gocv.IMWrite("debug/matches.png", outImage)
}
