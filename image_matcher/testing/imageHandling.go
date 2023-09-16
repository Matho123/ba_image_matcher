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

	//image := gocv.IMRead(filePath, gocv.IMReadGrayScale)
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal("Error opening the image: ", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		log.Fatal("Error decoding the image: ", err)
	}

	log.Println("successfully loaded: ", filePath)
	return &service.RawImage{ExternalReference: filePath, Data: img}
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
