package testing

import (
	"image_matcher/image-handling"
	"image_matcher/service"
	"log"
)

func populateDatabase([]string) {
	paths := image_handling.GetFilePathsFromDirectory("images/originals")

	var chunkSize = 10

	for index1 := chunkSize; index1 <= 800; index1 += chunkSize {
		rawImages := image_handling.LoadImagesFromDirectory(paths[(index1 - chunkSize):(index1)])

		//register db set images
		err := service.AnalyzeAndSaveDatabaseImage(rawImages)
		if err != nil {
			log.Println("Error while analysing and saving db images: ", err)
		}

		for _, rawImage := range rawImages {
			if index1 == 10 || index1 == 130 || index1 == 420 || index1 == 730 {
				service.InsertSearchImage(*rawImage, "identical")
			}
			if index1 == 70 || index1 == 130 || index1 == 420 || index1 == 750 {
				service.InsertSearchImage(*rawImage, "scaled")
			}
			if index1 == 10 || index1 == 180 || index1 == 750 {
				service.InsertSearchImage(*rawImage, "rotated")
			}
			if index1 == 70 || index1 == 180 || index1 == 540 || index1 == 730 {
				service.InsertSearchImage(*rawImage, "mirrored")
			}
			if index1 == 50 || index1 == 280 || index1 == 540 || index1 == 780 {
				service.InsertSearchImage(*rawImage, "background")
			}
			if index1 == 10 || index1 == 300 || index1 == 670 || index1 == 780 {
				service.InsertSearchImage(*rawImage, "moved")
			}
			if index1 == 10 || index1 == 300 || index1 == 650 || index1 == 710 {
				service.InsertSearchImage(*rawImage, "part")
			}
		}

		rawImages = nil
	}

	//create uniques for search sets
	for index := 800 + chunkSize; index <= 800+len(paths[800:]); index += chunkSize {
		rawImages := image_handling.LoadImagesFromDirectory(paths[(index - chunkSize):(index)])

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
	paths := image_handling.GetFilePathsFromDirectory("images/originals")

	log.Println(len(paths[800:]))
	var chunkSize = 10

	for index := 800 + chunkSize; index <= 800+len(paths[800:]); index += chunkSize {
		rawImages := image_handling.LoadImagesFromDirectory(paths[(index - chunkSize):(index)])

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
