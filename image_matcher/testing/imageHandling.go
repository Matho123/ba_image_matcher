package testing

import (
	"gocv.io/x/gocv"
	"image-matcher/image_matcher/service"
	_ "image-matcher/image_matcher/service"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var allowedImageExtensions = [...]string{".png", ".jpg"}

func loadImagesFromPath(filePath string) []*service.RawImage {
	fileInfo, err := os.Stat(filePath)

	if err != nil {
		log.Println(err)
		return []*service.RawImage{}
	}

	if fileInfo.IsDir() {
		return loadImagesFromDirectory(filePath)
	} else {
		return []*service.RawImage{loadImage(filePath)}
	}

}

func loadImagesFromDirectory(directoryPath string) []*service.RawImage {
	var rawImageDtos []*service.RawImage

	err := filepath.Walk(directoryPath, func(filePath string, fileInfo fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		var rawImageDto *service.RawImage

		if !fileInfo.IsDir() {
			rawImageDto = loadImage(filePath)
		}

		if rawImageDto != nil {
			rawImageDtos = append(rawImageDtos, rawImageDto)
		}
		return nil
	})

	if err != nil {
		log.Println(err)
	}
	return rawImageDtos
}

func loadImage(filePath string) *service.RawImage {
	if !isAllowedImageFile(filePath) {
		return nil
	}
	image := gocv.IMRead(filePath, 0)

	if image.Empty() {
		return nil
	}
	log.Println("successfully loaded: " + filePath)
	return &service.RawImage{ExternalReference: filePath, Data: image}
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
