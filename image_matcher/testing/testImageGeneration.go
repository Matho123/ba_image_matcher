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
		originals := image_handling.LoadImagesFromDirectory(paths[(index1 - chunkSize):(index1)])

		//register db set images
		err := service.AnalyzeAndSaveDatabaseImage(originals)
		if err != nil {
			log.Println("Error while analysing and saving db images: ", err)
		}

		for _, original := range originals {
			if index1 == 10 || index1 == 130 || index1 == 420 || index1 == 730 {
				modifier := "identical"
				variations := image_handling.GenerateDuplicateVariations(original, modifier)
				service.InsertDuplicateSearchImage(variations, original.ExternalReference, modifier)
			}
			if index1 == 70 || index1 == 130 || index1 == 420 || index1 == 750 {
				modifier := image_handling.SCALED
				variations := image_handling.GenerateDuplicateVariations(original, modifier)
				service.InsertDuplicateSearchImage(variations, original.ExternalReference, modifier)
			}
			if index1 == 10 || index1 == 180 || index1 == 750 {
				modifier := image_handling.ROTATED
				variations := image_handling.GenerateDuplicateVariations(original, modifier)
				service.InsertDuplicateSearchImage(variations, original.ExternalReference, modifier)
			}
			if index1 == 70 || index1 == 180 || index1 == 540 || index1 == 730 {
				modifier := image_handling.MIRRORED
				variations := image_handling.GenerateDuplicateVariations(original, modifier)
				service.InsertDuplicateSearchImage(variations, original.ExternalReference, modifier)
			}
			if index1 == 50 || index1 == 280 || index1 == 540 || index1 == 780 {
				modifier := image_handling.BACKGROUND
				variations := image_handling.GenerateDuplicateVariations(original, modifier)
				service.InsertDuplicateSearchImage(variations, original.ExternalReference, modifier)
			}
			if index1 == 10 || index1 == 300 || index1 == 670 || index1 == 780 {
				modifier := image_handling.MOVED
				variations := image_handling.GenerateDuplicateVariations(original, modifier)
				service.InsertDuplicateSearchImage(variations, original.ExternalReference, modifier)
			}
			if index1 == 10 || index1 == 300 || index1 == 650 || index1 == 710 {
				modifier := image_handling.PART
				variations := image_handling.GenerateDuplicateVariations(original, modifier)
				service.InsertDuplicateSearchImage(variations, original.ExternalReference, modifier)
			}
			if index1 == 10 || index1 == 300 || index1 == 650 || index1 == 710 {
				variation := image_handling.GenerateMixedVariation(original)
				service.InsertDuplicateSearchImage(
					&[]image_handling.ImageVariation{*variation},
					original.ExternalReference,
					"mixed",
				)
			}
		}

		originals = nil
	}

	//create uniques for search sets
	for index := 800 + chunkSize; index <= 800+len(paths[800:]); index += chunkSize {
		originals := image_handling.LoadImagesFromDirectory(paths[(index - chunkSize):(index)])

		for _, original := range originals {
			service.GenerateAndInsertUniqueSearchImages(original)
		}

		originals = nil
	}

	paths = nil
}