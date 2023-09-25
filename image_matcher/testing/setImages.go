package testing

import (
	"image-matcher/image_matcher/service"
	"log"
)

func populateDatabase([]string) {
	paths := getFilePathsFromDirectory("images/originals")

	log.Println(len(paths))
	var chunkSize = 10

	for index := chunkSize; index <= 800; index += chunkSize {
		rawImages := loadImagesFromDirectory(paths[(index - chunkSize):(index)])

		//register db set images
		err := service.AnalyzeAndSaveDatabaseImage(rawImages)
		if err != nil {
			log.Println("Error while analysing and saving db images: ", err)
		}

		rawImages = nil
	}

	//create uniques for search sets
	for index := 800 + chunkSize; index <= 800+len(paths[800:]); index += chunkSize {
		rawImages := loadImagesFromDirectory(paths[(index - chunkSize):(index)])

		for _, rawImage := range rawImages {
			service.GenerateUnique(*rawImage, "identical")
			service.GenerateUnique(*rawImage, "scaled")
			service.GenerateUnique(*rawImage, "rotated")
			service.GenerateUnique(*rawImage, "mirrored")
			service.GenerateUnique(*rawImage, "background")
			service.GenerateUnique(*rawImage, "moved")
			service.GenerateUnique(*rawImage, "part")
		}

		rawImages = nil
	}

	paths = nil
}

func pop([]string) {
	paths := getFilePathsFromDirectory("images/originals")

	log.Println(len(paths[800:]))
	var chunkSize = 10

	for index := 800 + chunkSize; index <= 800+len(paths[800:]); index += chunkSize {
		rawImages := loadImagesFromDirectory(paths[(index - chunkSize):(index)])

		for _, rawImage := range rawImages {
			service.GenerateUnique(*rawImage, "identical")
			service.GenerateUnique(*rawImage, "scaled")
			service.GenerateUnique(*rawImage, "rotated")
			service.GenerateUnique(*rawImage, "mirrored")
			service.GenerateUnique(*rawImage, "background")
			service.GenerateUnique(*rawImage, "moved")
			service.GenerateUnique(*rawImage, "part")
		}

		rawImages = nil
	}

	paths = nil
}
