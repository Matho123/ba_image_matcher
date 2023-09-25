package testing

import (
	"image"
	"image-matcher/image_matcher/service"
	_ "image-matcher/image_matcher/service"
	_ "image/jpeg"
	_ "image/png"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var allowedImageExtensions = [...]string{".png", ".jpg"}

func loadImagesFromPath(path string) []*service.RawImage {
	fileInfo, err := os.Stat(path)

	if err != nil {
		log.Println(err)
		return []*service.RawImage{}
	}

	if fileInfo.IsDir() {
		paths := getFilePathsFromDirectory(path)
		return loadImagesFromDirectory(paths)
	} else {
		return []*service.RawImage{loadImage(path)}
	}

}

func getFilePathsFromDirectory(directoryPath string) []string {
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

func loadImagesFromDirectory(filePaths []string) []*service.RawImage {
	var rawImageDtos []*service.RawImage

	for _, path := range filePaths {
		rawImageDto := loadImage(path)
		if rawImageDto != nil {
			rawImageDtos = append(rawImageDtos, rawImageDto)
		}
	}

	return rawImageDtos
}

func loadImage(path string) *service.RawImage {
	if !isAllowedImageFile(path) {
		return nil
	}

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

	filenameWithExt := filepath.Base(path)
	filenameWithoutExt := strings.TrimSuffix(filenameWithExt, filepath.Ext(filenameWithExt))

	return &service.RawImage{ExternalReference: filenameWithoutExt, Data: img}
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
