package image_handling

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
	var rawImage []*RawImage

	for _, path := range filePaths {
		rawImageDto := LoadRawImage(path)
		if rawImageDto != nil {
			rawImage = append(rawImage, rawImageDto)
		}
	}

	return rawImage
}

func LoadRawImage(path string) *RawImage {
	if !isAllowedImageFile(path) {
		return nil
	}

	img := loadImageFromDisk(path)

	filenameWithExt := filepath.Base(path)
	filenameWithoutExt := strings.TrimSuffix(filenameWithExt, filepath.Ext(filenameWithExt))

	return &RawImage{ExternalReference: filenameWithoutExt, Data: *img}
}

func loadImageFromDisk(path string) *image.Image {
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

func SaveImageToDisk(path string, image image.Image) {
	newPath := path + ".png"
	outputFile, err := os.Create(newPath)
	if err != nil {
		log.Println("Error while creating outputfile for image: ", err)
		return
	}
	defer outputFile.Close()

	err = png.Encode(outputFile, image)
	if err != nil {
		log.Println("Error while saving image "+path+" to disk: ", err)
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

func DrawMatches(
	image1 *gocv.Mat,
	keypoints1 []gocv.KeyPoint,
	image2 *gocv.Mat,
	keypoints2 []gocv.KeyPoint,
	bestMatches *[]gocv.DMatch,
) {
	if bestMatches == nil {
		log.Println("Can't draw debug matches because there are no matches")
		return
	}
	outImage := gocv.NewMat()
	gocv.DrawMatches(
		*image1,
		keypoints1,
		*image2,
		keypoints2,
		*bestMatches,
		&outImage,
		color.RGBA{R: 255, A: 100},
		color.RGBA{R: 255},
		[]byte{},
		gocv.DrawMatchesFlag(0),
	)
	gocv.IMWrite("debug/matches.png", outImage)
}
