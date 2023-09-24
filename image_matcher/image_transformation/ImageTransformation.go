package image_transformation

import (
	"github.com/disintegration/imaging"
	"image"
	"image/color"
	"image/jpeg"
	"log"
	"math/rand"
	"os"
)

func ResizeImage(img *image.Image) (image.Image, uint8) {
	scalingFactors := []int{2, 4, 10}
	randomIndex := rand.Intn(len(scalingFactors))
	scalingFactor := scalingFactors[randomIndex]

	newWidth := (*img).Bounds().Dx() / scalingFactor
	newHeight := (*img).Bounds().Dy() / scalingFactor

	scaled := imaging.Resize(*img, newWidth, newHeight, imaging.Lanczos)
	return scaled, uint8(scalingFactor)
}

func RotateImage(img *image.Image) (image.Image, int) {
	angle := rand.Intn(360)

	rotatedImage := imaging.Rotate(*img, float64(angle), color.Transparent)

	return rotatedImage, angle
}

func MirrorImage(img *image.Image) (image.Image, string) {
	horizontal := rand.Intn(2) == 0
	var mirroredImage image.Image
	var axis string

	if horizontal {
		mirroredImage = imaging.FlipH(*img)
		axis = "Y-Axis"
	} else {
		mirroredImage = imaging.FlipV(*img)
		axis = "X-Axis"
	}

	return mirroredImage, axis
}

func ChangeBackgroundColor(img *image.Image) (image.Image, color.Color) {
	r := uint8(rand.Intn(255))
	g := uint8(rand.Intn(255))
	b := uint8(rand.Intn(255))
	newBackground := color.RGBA{R: r, G: g, B: b, A: 255}

	newImage := imaging.New((*img).Bounds().Size().X, (*img).Bounds().Size().Y, newBackground)
	newImage = imaging.Overlay(newImage, *img, image.Pt(0, 0), 1.0)

	return newImage, newBackground
}

func MoveMotive(img *image.Image) (image.Image, int) {
	croppedImage := cropImage(img)
	originalWidth := (*img).Bounds().Dx()
	originalHeight := (*img).Bounds().Dy()

	croppedWidth := croppedImage.Bounds().Dx()
	croppedHeight := croppedImage.Bounds().Dy()

	maxX := (originalWidth - croppedWidth) / 2
	maxY := (originalHeight - croppedHeight) / 2

	movedX := rand.Intn(maxX)
	movedY := rand.Intn(maxY)

	newImage := imaging.New(originalWidth, originalHeight, color.Transparent)

	newImage = imaging.Paste(newImage, croppedImage, image.Point{X: movedX, Y: movedY})

	return newImage, movedX + movedY
}

func IntegrateInOtherImage(img *image.Image) {

}

func cropImage(img *image.Image) image.Image {
	var minX, minY, maxX, maxY int

	for y := 0; y < (*img).Bounds().Dy(); y++ {
		for x := 0; x < (*img).Bounds().Dx(); x++ {
			_, _, _, alpha := (*img).At(x, y).RGBA()
			if alpha == 0 {
				continue
			}
			if x < minX || minX == 0 {
				minX = x
			}
			if x > maxX {
				maxX = x
			}
			if y < minY || minY == 0 {
				minY = y
			}
			if y > maxY {
				maxY = y
			}
		}
	}

	croppedImage := imaging.Crop(*img, image.Rect(minX, minY, maxX, minY))

	return croppedImage
}

func SaveImageToDisk(name string, image image.Image) {
	outputFile, err := os.Create("images/" + name + ".jpg")
	if err != nil {
		log.Println("Error while creating outputfile for image: ", err)
		return
	}
	defer outputFile.Close()

	err = jpeg.Encode(outputFile, image, nil)
	if err != nil {
		log.Println("Error while saving image to disk: ", err)
		return
	}
}

//func ChangeMotiveColor(img *image.Image) image.Image {
//
//}
//
//func createMotiveMask(img *image.Image) *image.Gray {
//	bounds := (*img).Bounds()
//	motiveMask := image.NewGray(bounds)
//
//	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
//		for x := bounds.Min.X; x < bounds.Max.X; x++ {
//			pixelColor := (*img).At(x, y).(color.NRGBA)
//
//		}
//	}
//	return motiveMask
//}
//
//func calculateColorDistance(color1, color2 color.NRGBA) float64 {
//	red1, green1, blue1 := color1.R, color1.G, color1.B
//	red2, green2, blue2 := color2.R, color2.G, color2.B
//
//	deltaRed := float64(red1 - red2)
//	deltaGreen := float64(green1 - green2)
//	deltaBlue := float64(blue1 - blue2)
//
//	return math.Sqrt(deltaRed*deltaRed + deltaGreen*deltaGreen + deltaBlue*deltaBlue)
//}
