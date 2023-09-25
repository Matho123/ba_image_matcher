package testing

import (
	"image-matcher/image_matcher/service"
	"log"
)

func populateDatabase([]string) {
	paths := getFilePathsFromDirectory("images/originals")

	var chunkSize = 10

	for index := chunkSize; index < 800; index += chunkSize {
		rawImages := loadImagesFromDirectory(paths[(index - chunkSize):(index - 1)])

		//register db set images
		err := service.AnalyzeAndSaveDatabaseImage(rawImages)
		if err != nil {
			log.Println("Error while analysing and saving db images: ", err)
		}

		rawImages = nil
	}

	//create uniques for search sets

	for index := 800 + chunkSize; index < len(paths[800:]); index += chunkSize {
		rawImages := loadImagesFromDirectory(paths[(index - chunkSize):(index - 1)])

		for _, rawImage := range rawImages {
			service.InsertSearchImage(*rawImage, "identical", false)
			service.InsertSearchImage(*rawImage, "scaled", false)
			service.InsertSearchImage(*rawImage, "rotated", false)
			service.InsertSearchImage(*rawImage, "mirrored", false)
			service.InsertSearchImage(*rawImage, "background", false)
			service.InsertSearchImage(*rawImage, "moved", false)
			service.InsertSearchImage(*rawImage, "part", false)
		}

		rawImages = nil
	}

	paths = nil
}

func insertData(rawImages []*service.RawImage) {

	//create duplicates for searchset
	for index, rawImage := range rawImages[0:800] {
		if index >= 0 && index < 200 {
			service.InsertSearchImage(*rawImage, "identical", true)
		}
		if index >= 100 && index < 300 {
			service.InsertSearchImage(*rawImage, "scaled", true)
		}
		if index >= 200 && index < 400 {
			service.InsertSearchImage(*rawImage, "rotated", true)
		}
		if index >= 300 && index < 500 {
			service.InsertSearchImage(*rawImage, "mirrored", true)
		}
		if index >= 400 && index < 600 {
			service.InsertSearchImage(*rawImage, "background", true)
		}
		if index >= 500 && index < 700 {
			service.InsertSearchImage(*rawImage, "moved", true)
		}
		if index >= 600 && index < 800 {
			service.InsertSearchImage(*rawImage, "part", true)
		}
	}
}
