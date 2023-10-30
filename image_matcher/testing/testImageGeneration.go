package testing

import (
	"database/sql"
	"fmt"
	"image_matcher/image_analyzer"
	"image_matcher/image_database"
	"image_matcher/image_handling"
	"image_matcher/image_service"
	"log"
)

func populateDatabase(directoryPath string) {
	paths := image_handling.GetFilePathsFromDirectory(directoryPath)
	var chunkSize = 10

	//creates duplicates for search set
	offset := 0
	for {
		limit := offset + chunkSize

		if limit > len(paths) {
			limit = len(paths)
		}

		originals := image_handling.LoadImagesFromDirectory(paths[offset:limit])

		for _, original := range originals {
			for _, modifier := range image_handling.Modifiers {
				variations := image_handling.GenerateDuplicateVariations(original, modifier)
				image_service.InsertDuplicateSearchImage(variations, original.ExternalReference, modifier)
				variations = nil
			}
			variation := image_handling.GenerateMixedVariation(original)
			image_service.InsertDuplicateSearchImage(
				&[]image_handling.ImageVariation{*variation},
				original.ExternalReference,
				"mixed",
			)
			variation = nil
		}

		if len(originals) < chunkSize {
			break
		}
		offset += chunkSize
	}
}

// create uniques for search sets
func generateUniques(directoryPath string) {

	paths := image_handling.GetFilePathsFromDirectory(directoryPath)
	var chunkSize = 10

	for index := 0; index <= len(paths); index += chunkSize {
		limit := index + chunkSize
		if limit > len(paths) {
			limit = len(paths)
		}
		originals := image_handling.LoadImagesFromDirectory(paths[index:limit])

		for _, original := range originals {
			image_service.GenerateAndInsertUniqueSearchImages(original)
		}

		originals = nil
	}

	paths = nil
}

func updateDatabaseWithNewHash([]string) {
	err := image_database.ApplyDatabaseOperation(func(databaseConnection *sql.DB) {
		references, err := image_database.GetForbiddenReferences(databaseConnection)
		if err != nil {
			log.Println(err)
		}
		println(len(*references))
		for _, reference := range *references {
			rawImage := image_handling.LoadRawImage(fmt.Sprintf("images/originals/%s.png", reference))
			hash, _ := image_analyzer.CalculateOrientedPHash(&rawImage.Data)
			err := image_database.InsertRotationHashIntoDatabase(databaseConnection, reference, hash)
			if err != nil {
				log.Println(err)
			}
		}

	})
	if err != nil {
		log.Println(err)
	}

}
