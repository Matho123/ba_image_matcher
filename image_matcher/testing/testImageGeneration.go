package testing

import (
	"image_matcher/image-handling"
	"image_matcher/service"
	"log"
)

var indizes = []int{10, 50, 70, 130, 180, 280, 300, 420, 540, 650, 670, 710, 730, 750, 780}

func contains(arr *[]int, number int) bool {
	for _, e := range *arr {
		if number == e {
			return true
		}
	}
	return false
}

func populateDatabase([]string) {

	paths := image_handling.GetFilePathsFromDirectory("images/originals")

	arr3 := indizes[9:13]
	arr4 := indizes[6:11]
	arr6 := indizes[8:15]
	arr12 := indizes[0:12]

	var chunkSize = 10

	for index1 := chunkSize; index1 <= 850; index1 += chunkSize {
		originals := image_handling.LoadImagesFromDirectory(paths[(index1 - chunkSize):(index1)])

		//register db set images
		err := service.AnalyzeAndSaveDatabaseImage(originals)
		if err != nil {
			log.Println("Error while analysing and saving db images: ", err)
		}

		for _, original := range originals {
			if contains(&arr12, index1) {
				modifier := "identical"
				variations := image_handling.GenerateDuplicateVariations(original, modifier)
				service.InsertDuplicateSearchImage(variations, original.ExternalReference, modifier)
				variations = nil
			}
			if contains(&arr4, index1) {
				modifier := image_handling.SCALED
				variations := image_handling.GenerateDuplicateVariations(original, modifier)
				service.InsertDuplicateSearchImage(variations, original.ExternalReference, modifier)
				variations = nil
			}
			if contains(&arr3, index1) {
				modifier := image_handling.ROTATED
				variations := image_handling.GenerateDuplicateVariations(original, modifier)
				service.InsertDuplicateSearchImage(variations, original.ExternalReference, modifier)
				variations = nil
			}
			if contains(&arr6, index1) {
				modifier := image_handling.MIRRORED
				variations := image_handling.GenerateDuplicateVariations(original, modifier)
				service.InsertDuplicateSearchImage(variations, original.ExternalReference, modifier)
				variations = nil
			}
			if contains(&arr12, index1) {
				modifier := image_handling.BACKGROUND
				variations := image_handling.GenerateDuplicateVariations(original, modifier)
				service.InsertDuplicateSearchImage(variations, original.ExternalReference, modifier)
				variations = nil
			}
			if contains(&arr12, index1) {
				modifier := image_handling.MOVED
				variations := image_handling.GenerateDuplicateVariations(original, modifier)
				service.InsertDuplicateSearchImage(variations, original.ExternalReference, modifier)
				variations = nil
			}
			if contains(&arr12, index1) {
				modifier := image_handling.PART
				variations := image_handling.GenerateDuplicateVariations(original, modifier)
				service.InsertDuplicateSearchImage(variations, original.ExternalReference, modifier)
				variations = nil
			}
			if contains(&arr12, index1) {
				variation := image_handling.GenerateMixedVariation(original)
				service.InsertDuplicateSearchImage(
					&[]image_handling.ImageVariation{*variation},
					original.ExternalReference,
					"mixed",
				)
				variation = nil
			}
		}

		originals = nil
	}

	//create uniques for search sets
	for index := 850 + chunkSize; index <= 850+len(paths[850:]); index += chunkSize {
		originals := image_handling.LoadImagesFromDirectory(paths[(index - chunkSize):(index)])

		for _, original := range originals {
			service.GenerateAndInsertUniqueSearchImages(original)
		}

		originals = nil
	}

	paths = nil
}
