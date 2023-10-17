package service

import (
	"gocv.io/x/gocv"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

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

	log.Println("successfully loaded: ", path)
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

func ConvertImageMatToByteArray(image gocv.Mat) []byte {
	if image.Empty() {
		log.Println("descriptor is empty!")
		return nil
	}

	nativeByteBuffer, err := gocv.IMEncode(".png", image)
	if err != nil {
		log.Println("unable to convert image to gocv.NativeByteBuffer! ", err)
		return nil
	}
	image.ToBytes()
	return nativeByteBuffer.GetBytes()
}

func ConvertByteArrayToMat(imageBytes []byte) gocv.Mat {
	imageMat, err := gocv.IMDecode(imageBytes, -1)

	if err != nil || imageMat.Empty() {
		log.Println("unable to convert bytes to gocv.mat")
	}
	return imageMat
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
