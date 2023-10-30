package image_service

import (
	"database/sql"
	"fmt"
	"image_matcher/image_database"
	"image_matcher/image_handling"
	"log"
)

const (
	IDENTICAL  = "identical"
	SCALED     = "scaled"
	ROTATED    = "rotated"
	MIRRORED   = "mirrored"
	MOVED      = "moved"
	BACKGROUND = "background"
	PART       = "part"
	MIXED      = "mixed"
)

var Scenarios = []string{IDENTICAL, SCALED, ROTATED, MIRRORED, MOVED, BACKGROUND, PART, MIXED}

func GetSearchImages(scenario string) *[]image_database.SearchImageEntity {
	var searchSetImages []image_database.SearchImageEntity

	err := image_database.ApplyDatabaseOperation(func(databaseConnection *sql.DB) {
		offset := 0
		for {
			retrievedImages, err := image_database.RetrieveChunkFromSearchSet(
				databaseConnection,
				scenario,
				offset,
				image_database.MaxChunkSize+1,
			)
			if err != nil {
				log.Println("Error while retrieving chunk from search images: ", err)
			}

			numberOfRetrievedImages := len(retrievedImages)
			if numberOfRetrievedImages > 0 {
				searchSetImages = append(searchSetImages, retrievedImages[:numberOfRetrievedImages]...)
			}

			if len(retrievedImages) < image_database.MaxChunkSize+1 {
				break
			}
			offset += image_database.MaxChunkSize
		}
	})
	if err != nil {
		log.Println("Error while retrieving chunk from search images: ", err)
		return nil
	}

	return &searchSetImages
}

func InsertDuplicateSearchImage(variations *[]image_handling.ImageVariation, originalReference string, scenario string) {
	var externalReference = fmt.Sprintf("%s-%s", originalReference, scenario)
	var err error

	err = image_database.ApplyDatabaseOperation(func(databaseConnection *sql.DB) {
		for _, variation := range *variations {
			imageReference := externalReference + "-" + variation.ModificationInfo

			image_handling.SaveImageToDisk(fmt.Sprintf("images/variations/%s/%s", scenario, imageReference), variation.ModifiedImage)

			err = image_database.InsertImageIntoSearchSet(
				databaseConnection,
				image_database.SearchImageCreation{
					ExternalReference: imageReference,
					OriginalReference: originalReference,
					Scenario:          scenario,
					ModificationInfo:  variation.ModificationInfo,
				},
			)
			if err != nil {
				log.Println("failed to insert ", externalReference, err)
			}
		}
	})
	if err != nil {
		log.Println("Failed to open db for searchImages: ", err)
	}
}

func GenerateAndInsertUniqueSearchImages(originalImage *image_handling.RawImage) {
	err := image_database.ApplyDatabaseOperation(func(databaseConnection *sql.DB) {
		for _, scenario := range Scenarios {
			var variation *image_handling.ImageVariation
			if scenario == MIXED {
				variation = image_handling.GenerateMixedVariation(originalImage)
			} else {
				variation = image_handling.GenerateUniqueVariation(originalImage, scenario)
			}

			insertUniqueSearchImage(
				databaseConnection,
				variation,
				originalImage.ExternalReference,
				scenario,
			)
		}
	})
	if err != nil {
		log.Println("Failed to open db for searchImages: ", err)
	}
}

func insertUniqueSearchImage(
	databaseConnection *sql.DB,
	uniqueVariation *image_handling.ImageVariation,
	originalReference string,
	scenario string,
) {
	externalReference := fmt.Sprintf("%s-%s-%s", originalReference, scenario, uniqueVariation.ModificationInfo)

	image_handling.SaveImageToDisk(fmt.Sprintf("images/variations/%s/%s", scenario, externalReference), uniqueVariation.ModifiedImage)

	err := image_database.InsertImageIntoSearchSet(
		databaseConnection,
		image_database.SearchImageCreation{
			ExternalReference: externalReference,
			OriginalReference: "",
			Scenario:          scenario,
			ModificationInfo:  uniqueVariation.ModificationInfo,
		},
	)
	if err != nil {
		log.Println("failed to insert ", externalReference, err)
	}
}
