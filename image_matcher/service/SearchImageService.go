package service

import (
	"fmt"
	"image"
	"image-matcher/image_matcher/image_transformation"
	"log"
	"strconv"
)

type ModifiedImage struct {
	externalReference string
	originalReference string
	scenario          string
	notes             string
}

func GetSearchImages(scenario string) ([]SearchSetImage, error) {
	databaseConnection, err := openDatabaseConnection()
	if err != nil {
		return []SearchSetImage{}, err
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
	return searchSetImages, nil
}

func InsertSearchImage(originalImage RawImage, scenario string, hasOriginal bool) {
	databaseConnection, err := openDatabaseConnection()
	if err != nil {
		log.Println("Failed to open db for searchImages: ", err)
	}
	defer databaseConnection.Close()

	var externalReference = fmt.Sprintf("%s-%s", originalImage.ExternalReference, scenario)
	var variation image.Image
	var notes string

	switch scenario {
	case "scaled":
		scaled, scalingFactor := image_transformation.ResizeImage(&originalImage.Data)
		variation = scaled
		notes = strconv.Itoa(int(scalingFactor))
		break
	case "rotated":
		rotated, angle := image_transformation.RotateImage(&originalImage.Data)
		variation = rotated
		notes = string(rune(angle))
		break
	case "mirrored":
		mirrored, axis := image_transformation.MirrorImage(&originalImage.Data)
		variation = mirrored
		notes = axis
		break
	case "moved":
		moved, distance := image_transformation.MoveMotive(&originalImage.Data)
		variation = moved
		notes = string(rune(distance))
		break
	case "background":
		changed, bg := image_transformation.ChangeBackgroundColor(&originalImage.Data)
		variation = changed
		r, g, b, _ := bg.RGBA()
		notes = fmt.Sprintf("%d, %d, %d", r, g, b)
		break
	case "motive":
		break
	case "part":
		newImage, distance := image_transformation.IntegrateInOtherImage(&originalImage.Data)
		variation = newImage
		notes = string(rune(distance))
		break
	default:
		variation = originalImage.Data
		break
	}

	image_transformation.SaveImageToDisk(scenario+"/"+externalReference, variation)

	originalReference := ""
	if hasOriginal {
		originalReference = originalImage.ExternalReference
	}

	err = insertImageIntoSearchSet(
		databaseConnection,
		ModifiedImage{
			externalReference: externalReference,
			originalReference: originalReference,
			scenario:          scenario,
			notes:             notes,
		},
	)
	if err != nil {
		println("failed to insert ", externalReference, err)
	}
}
