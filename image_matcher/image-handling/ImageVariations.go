package image_handling

import (
	"fmt"
	"image"
	"math/rand"
	"strconv"
)

const (
	SCALED     = "scaled"
	ROTATED    = "rotated"
	MIRRORED   = "mirrored"
	MOVED      = "moved"
	BACKGROUND = "background"
	PART       = "part"
)

type ImageVariation struct {
	Img   image.Image
	Notes string
}

var scalingFactors = []int{2, 4, 10}

var rotationAngles = []float64{5, 10, 45, 90, 180}

func GenerateVariations(originalImage *RawImage, modifier string) *[]ImageVariation {
	var variations *[]ImageVariation

	switch modifier {
	case SCALED:
		variations = generateScaledVariations(&originalImage.Data)
		break
	case ROTATED:
		variations = generateRotatedVariations(&originalImage.Data)
		break
	case MIRRORED:
		mirroredHorizontal, axisHorizontal := MirrorImage(&originalImage.Data, true)
		horizontalVariation := ImageVariation{mirroredHorizontal, axisHorizontal}

		mirroredVertical, axisVertical := MirrorImage(&originalImage.Data, false)
		verticalVariation := ImageVariation{mirroredVertical, axisVertical}
		variations = &[]ImageVariation{
			horizontalVariation,
			verticalVariation,
		}
		break
	case MOVED:
		moved, distance := MoveMotive(&originalImage.Data)
		variations = &[]ImageVariation{{moved, fmt.Sprintf("%.0f", distance)}}
		break
	case BACKGROUND:
		changed, bg := ChangeBackgroundColor(&originalImage.Data)
		r, g, b, _ := bg.RGBA()
		r8, g8, b8 := uint8(r>>8), uint8(g>>8), uint8(b>>8)
		variations = &[]ImageVariation{{changed, fmt.Sprintf("%d/%d/%d", r8, g8, b8)}}
		break
	case PART:
		newImage, distance := IntegrateInOtherImage(&originalImage.Data)
		variations = &[]ImageVariation{{Img: newImage, Notes: fmt.Sprintf("%.0f", distance)}}
		break
	default:
		variations = &[]ImageVariation{{Img: originalImage.Data, Notes: ""}}
		break
	}

	return variations
}

func GenerateUniqueVariation(originalImage *RawImage, modifier string) *ImageVariation {
	var variation image.Image
	var notes string

	switch modifier {
	case SCALED:
		randomIndex := rand.Intn(len(scalingFactors))
		scalingFactor := scalingFactors[randomIndex]

		scaled := ResizeImage(&originalImage.Data, scalingFactor)
		variation = scaled
		notes = strconv.Itoa(scalingFactor)
		break
	case ROTATED:
		randomIndex := rand.Intn(len(rotationAngles))
		angle := rotationAngles[randomIndex]

		rotated := RotateImage(&originalImage.Data, angle)
		variation = rotated
		notes = fmt.Sprintf("%.0f", angle)
		break
	case "mirrored":
		horizontal := rand.Intn(2) == 0

		mirrored, axis := MirrorImage(&originalImage.Data, horizontal)
		variation = mirrored
		notes = axis
		break
	case "moved":
		moved, distance := MoveMotive(&originalImage.Data)
		variation = moved
		notes = fmt.Sprintf("%.0f", distance)
		break
	case "background":
		changed, bg := ChangeBackgroundColor(&originalImage.Data)
		variation = changed
		r, g, b, _ := bg.RGBA()
		r8, g8, b8 := uint8(r>>8), uint8(g>>8), uint8(b>>8)
		notes = fmt.Sprintf("%d, %d, %d", r8, g8, b8)
		break
	case "motive":
		break
	case "part":
		newImage, distance := IntegrateInOtherImage(&originalImage.Data)
		variation = newImage
		notes = fmt.Sprintf("%.0f", distance)
		break
	default:
		variation = originalImage.Data
		break
	}

	return &ImageVariation{
		Img:   variation,
		Notes: notes,
	}
}

func generateScaledVariations(img *image.Image) *[]ImageVariation {
	var variations []ImageVariation
	for _, scalingFactor := range scalingFactors {
		scaled := ResizeImage(img, scalingFactor)
		variations = append(
			variations,
			ImageVariation{
				Img:   scaled,
				Notes: strconv.Itoa(scalingFactor),
			},
		)
	}
	return &variations
}

func generateRotatedVariations(img *image.Image) *[]ImageVariation {
	var variations []ImageVariation
	for _, angle := range rotationAngles {
		rotated := RotateImage(img, angle)
		variations = append(
			variations,
			ImageVariation{
				Img:   rotated,
				Notes: fmt.Sprintf("%.0f", angle),
			},
		)
	}
	return &variations
}
