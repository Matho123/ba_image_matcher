package image_handling

import (
	"fmt"
	"image"
	"math/rand"
	"strconv"
	"time"
)

const (
	IDENTICAL  = "identical"
	SCALED     = "scaled"
	ROTATED    = "rotated"
	MIRRORED   = "mirrored"
	MOVED      = "moved"
	BACKGROUND = "background"
	PART       = "part"
)

var Modifiers = []string{IDENTICAL, SCALED, ROTATED, MIRRORED, MOVED, BACKGROUND, PART}

var scalingFactors = []int{2, 4, 10}

var rotationAngles = []float64{5, 10, 45, 90, 180}

type ImageVariation struct {
	ModifiedImage    image.Image
	ModificationInfo string
}

func GenerateDuplicateVariations(originalImage *RawImage, modifier string) *[]ImageVariation {

	switch modifier {
	case SCALED:
		return generateAllScaledVariations(&originalImage.Data)

	case ROTATED:
		return generateAllRotatedVariations(&originalImage.Data)

	case MIRRORED:
		return generateAllMirroredVariations(&originalImage.Data)

	default:
		modifiedImage, modificationInfo := modifyImage(&originalImage.Data, modifier)

		return &[]ImageVariation{{*modifiedImage, modificationInfo}}
	}

}

func GenerateUniqueVariation(originalImage *RawImage, modifier string) *ImageVariation {
	modifiedImage, modificationInfo := modifyImage(&originalImage.Data, modifier)

	return &ImageVariation{
		ModifiedImage:    *modifiedImage,
		ModificationInfo: modificationInfo,
	}
}

func GenerateMixedVariation(originalImage *RawImage) *ImageVariation {
	shuffledModifiers := shuffleArray(Modifiers)

	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	modifierAmount := random.Intn(4)

	modifiedImage := &originalImage.Data
	modificationInfo := ""
	for i := 0; i < modifierAmount; i++ {
		modifier := shuffledModifiers[i]
		modifiedImage, _ = modifyImage(modifiedImage, modifier)
		modificationInfo += modifier + "-"
	}
	return &ImageVariation{*modifiedImage, modificationInfo}
}

func shuffleArray(array []string) []string {
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := len(array) - 1; i > 0; i-- {
		j := random.Intn(i + 1)
		array[i], array[j] = array[j], array[i]
	}
	return array
}

func modifyImage(originalImage *image.Image, modifier string) (*image.Image, string) {
	random := rand.New(rand.NewSource(time.Now().UnixNano()))

	switch modifier {
	case SCALED:
		randomIndex := random.Intn(len(scalingFactors))
		scalingFactor := scalingFactors[randomIndex]

		scaled := ResizeImage(originalImage, scalingFactor)
		return &scaled, strconv.Itoa(scalingFactor)

	case ROTATED:
		randomIndex := random.Intn(len(rotationAngles))
		angle := rotationAngles[randomIndex]

		rotated := RotateImage(originalImage, angle)
		return &rotated, fmt.Sprintf("%.0f", angle)

	case MIRRORED:
		horizontal := random.Intn(2) == 0

		mirrored, axis := MirrorImage(originalImage, horizontal)
		return &mirrored, axis

	case MOVED:
		moved, distance := MoveMotive(originalImage)

		return &moved, fmt.Sprintf("%.0f", distance)

	case BACKGROUND:
		changed, bg := ChangeBackgroundColor(originalImage)
		r, g, b, _ := bg.RGBA()
		r8, g8, b8 := uint8(r>>8), uint8(g>>8), uint8(b>>8)

		return &changed, fmt.Sprintf("%d, %d, %d", r8, g8, b8)

	case PART:
		newImage, distance := IntegrateInOtherImage(originalImage)

		return &newImage, fmt.Sprintf("%.0f", distance)

	default:
		return originalImage, ""
	}
}

func generateAllScaledVariations(originalImage *image.Image) *[]ImageVariation {
	var variations []ImageVariation
	for _, scalingFactor := range scalingFactors {
		scaled := ResizeImage(originalImage, scalingFactor)
		variations = append(
			variations,
			ImageVariation{
				ModifiedImage:    scaled,
				ModificationInfo: strconv.Itoa(scalingFactor),
			},
		)
	}
	return &variations
}

func generateAllRotatedVariations(originalImage *image.Image) *[]ImageVariation {
	var variations []ImageVariation
	for _, angle := range rotationAngles {
		rotated := RotateImage(originalImage, angle)
		variations = append(
			variations,
			ImageVariation{
				ModifiedImage:    rotated,
				ModificationInfo: fmt.Sprintf("%.0f", angle),
			},
		)
	}
	return &variations
}

func generateAllMirroredVariations(originalImage *image.Image) *[]ImageVariation {
	mirroredHorizontal, axisHorizontal := MirrorImage(originalImage, true)
	horizontalVariation := ImageVariation{mirroredHorizontal, axisHorizontal}

	mirroredVertical, axisVertical := MirrorImage(originalImage, false)
	verticalVariation := ImageVariation{mirroredVertical, axisVertical}

	return &[]ImageVariation{
		horizontalVariation,
		verticalVariation,
	}
}
