package service

import (
	"database/sql"
	"fmt"
	"image_matcher/image-handling"
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

var SCENARIOS = []string{IDENTICAL, SCALED, ROTATED, MIRRORED, MOVED, BACKGROUND, PART, MIXED}

var scalingFactors = []int{2, 4, 10}

var rotationAngles = []float64{5, 10, 45, 90, 180}

func GetSearchImages(scenario string) *[]SearchSetImage {
	databaseConnection, err := openDatabaseConnection()
	if err != nil {
		log.Println("Error while retrieving chunk from search images: ", err)
		return nil
	}
	defer databaseConnection.Close()

	var searchSetImages []SearchSetImage

	offset := 0
	for {
		retrievedImages, err := retrieveChunkFromSearchSet(
			databaseConnection,
			scenario,
			offset,
			MaxChunkSize+1,
		)
		if err != nil {
			log.Println("Error while retrieving chunk from search images: ", err)
		}

		searchSetImages = append(searchSetImages, retrievedImages...)

		if len(retrievedImages) < MaxChunkSize+1 {
			break
		}
		offset += MaxChunkSize
	}
	return &searchSetImages
}

func InsertDuplicateSearchImage(variations *[]image_handling.ImageVariation, originalReference string, scenario string) {
	databaseConnection, err := openDatabaseConnection()
	if err != nil {
		log.Println("Failed to open db for searchImages: ", err)
	}
	defer databaseConnection.Close()

	var externalReference = fmt.Sprintf("%s-%s", originalReference, scenario)

	for _, variation := range *variations {
		imageReference := externalReference + "-" + variation.Notes

		image_handling.SaveImageToDisk(scenario+"/"+imageReference, variation.Img)

		err = insertImageIntoSearchSet(
			databaseConnection,
			ModifiedImage{
				externalReference: imageReference,
				originalReference: originalReference,
				scenario:          scenario,
				notes:             variation.Notes,
			},
		)
		if err != nil {
			log.Println("failed to insert ", externalReference, err)
		}
	}

}

func GenerateAndInsertUniqueSearchImages(originalImage *image_handling.RawImage) {
	databaseConnection, err := openDatabaseConnection()
	if err != nil {
		log.Println("Failed to open db for searchImages: ", err)
	}
	defer databaseConnection.Close()

	for _, scenario := range SCENARIOS {
		insertUniqueSearchImage(
			databaseConnection,
			image_handling.GenerateUniqueVariation(originalImage, scenario),
			originalImage.ExternalReference,
			scenario,
		)
	}
}

func insertUniqueSearchImage(
	databaseConnection *sql.DB,
	uniqueVariation *image_handling.ImageVariation,
	originalReference string,
	scenario string,
) {
	externalReference := fmt.Sprintf("%s-%s-%s", originalReference, scenario, uniqueVariation.Notes)

	image_handling.SaveImageToDisk(scenario+"/"+externalReference, uniqueVariation.Img)

	err := insertImageIntoSearchSet(
		databaseConnection,
		ModifiedImage{
			externalReference: externalReference,
			originalReference: "",
			scenario:          scenario,
			notes:             uniqueVariation.Notes,
		},
	)
	if err != nil {
		log.Println("failed to insert ", externalReference, err)
	}
}
