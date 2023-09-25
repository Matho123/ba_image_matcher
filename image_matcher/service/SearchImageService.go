package service

import (
	"fmt"
	"image"
	"image-matcher/image_matcher/image_transformation"
	"log"
	"math/rand"
	"strconv"
)

type ModifiedImage struct {
	externalReference string
	originalReference string
	scenario          string
	notes             string
}

type ImageVariation struct {
	img   image.Image
	notes string
}

var scalingFactors = []int{2, 4, 10}

var rotationAngles = []float64{5, 10, 45, 90, 180}

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

func InsertSearchImage(originalImage RawImage, scenario string) {
	databaseConnection, err := openDatabaseConnection()
	if err != nil {
		log.Println("Failed to open db for searchImages: ", err)
	}
	defer databaseConnection.Close()

	var externalReference = fmt.Sprintf("%s-%s", originalImage.ExternalReference, scenario)
	var variations []*ImageVariation

	switch scenario {
	case "scaled":
		variations = generateScaledVariations(&originalImage.Data)
		break
	case "rotated":
		variations = generateRotatedVariations(&originalImage.Data)
		break
	case "mirrored":
		mirroredHorizontal, axisHorizontal := image_transformation.MirrorImage(&originalImage.Data, true)
		horizontalVariation := ImageVariation{mirroredHorizontal, axisHorizontal}

		mirroredVertical, axisVertical := image_transformation.MirrorImage(&originalImage.Data, true)
		verticalVariation := ImageVariation{mirroredVertical, axisVertical}
		variations = []*ImageVariation{
			&horizontalVariation,
			&verticalVariation,
		}
		break
	case "moved":
		moved, distance := image_transformation.MoveMotive(&originalImage.Data)
		variations = []*ImageVariation{{moved, fmt.Sprintf("%.0f", distance)}}
		break
	case "background":
		changed, bg := image_transformation.ChangeBackgroundColor(&originalImage.Data)
		r, g, b, _ := bg.RGBA()
		r8, g8, b8 := uint8(r>>8), uint8(g>>8), uint8(b>>8)
		variations = []*ImageVariation{{changed, fmt.Sprintf("%d, %d, %d", r8, g8, b8)}}
		break
	case "motive":
		break
	case "part":
		newImage, distance := image_transformation.IntegrateInOtherImage(&originalImage.Data)
		variations = []*ImageVariation{{img: newImage, notes: fmt.Sprintf("%.0f", distance)}}
		break
	default:
		variations = []*ImageVariation{{img: originalImage.Data, notes: ""}}
		break
	}

	for _, variation := range variations {
		imageReference := externalReference + "-" + variation.notes

		image_transformation.SaveImageToDisk(scenario+"/"+imageReference, variation.img)

		err = insertImageIntoSearchSet(
			databaseConnection,
			ModifiedImage{
				externalReference: imageReference,
				originalReference: originalImage.ExternalReference,
				scenario:          scenario,
				notes:             variation.notes,
			},
		)
		if err != nil {
			log.Println("failed to insert ", externalReference, err)
		}
	}

}

func generateScaledVariations(img *image.Image) []*ImageVariation {
	var variations []*ImageVariation
	for _, scalingFactor := range scalingFactors {
		scaled := image_transformation.ResizeImage(img, scalingFactor)
		variations = append(
			variations,
			&ImageVariation{
				img:   scaled,
				notes: strconv.Itoa(scalingFactor),
			},
		)
	}
	return variations
}

func generateRotatedVariations(img *image.Image) []*ImageVariation {
	var variations []*ImageVariation
	for _, angle := range rotationAngles {
		rotated := image_transformation.RotateImage(img, angle)
		variations = append(
			variations,
			&ImageVariation{
				img:   rotated,
				notes: fmt.Sprintf("%.0f", angle),
			},
		)
	}
	return variations
}

func GenerateUnique(originalImage RawImage, scenario string) {
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
		randomIndex := rand.Intn(len(scalingFactors))
		scalingFactor := scalingFactors[randomIndex]

		scaled := image_transformation.ResizeImage(&originalImage.Data, scalingFactor)
		variation = scaled
		notes = strconv.Itoa(scalingFactor)
		break
	case "rotated":
		randomIndex := rand.Intn(len(rotationAngles))
		angle := rotationAngles[randomIndex]

		rotated := image_transformation.RotateImage(&originalImage.Data, angle)
		variation = rotated
		notes = fmt.Sprintf("%.0f", angle)
		break
	case "mirrored":
		horizontal := rand.Intn(2) == 0

		mirrored, axis := image_transformation.MirrorImage(&originalImage.Data, horizontal)
		variation = mirrored
		notes = axis
		break
	case "moved":
		moved, distance := image_transformation.MoveMotive(&originalImage.Data)
		variation = moved
		notes = fmt.Sprintf("%.0f", distance)
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
		notes = fmt.Sprintf("%.0f", distance)
		break
	default:
		variation = originalImage.Data
		break
	}

	externalReference = externalReference + "-" + notes

	image_transformation.SaveImageToDisk(scenario+"/"+externalReference, variation)

	err = insertImageIntoSearchSet(
		databaseConnection,
		ModifiedImage{
			externalReference: externalReference,
			originalReference: "",
			scenario:          scenario,
			notes:             notes,
		},
	)
	if err != nil {
		log.Println("failed to insert ", externalReference, err)
	}
}
