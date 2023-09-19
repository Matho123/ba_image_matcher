package image_transformation

import (
	"github.com/disintegration/imaging"
	"image"
	"image/color"
	"image/jpeg"
	"log"
	"os"
)

func ResizeImage(img *image.Image, width int, height int) image.Image {
	scaled := imaging.Resize(*img, width, height, imaging.Lanczos)
	saveImageToDisk("resized", scaled)
	return scaled
}

func RotateImage(img *image.Image, rotationAngle float64) image.Image {
	rotatedImage := imaging.Rotate(*img, rotationAngle, color.Transparent)
	saveImageToDisk("rotated", rotatedImage)
	return rotatedImage
}

func MirrorImage(img *image.Image, horizontal bool) image.Image {
	var mirroredImage image.Image
	if horizontal {
		mirroredImage = imaging.FlipH(*img)
	} else {
		mirroredImage = imaging.FlipV(*img)
	}
	saveImageToDisk("mirrored", mirroredImage)
	return mirroredImage
}

func ChangeBackgroundColor(img *image.Image, color color.Color) image.Image {
	newImage := imaging.New((*img).Bounds().Size().X, (*img).Bounds().Size().Y, color)
	newImage = imaging.Overlay(newImage, *img, image.Pt(0, 0), 1.0)
	saveImageToDisk("background", newImage)
	return newImage
}

func saveImageToDisk(name string, image image.Image) {
	outputFile, err := os.Create("images/" + name + ".jpg")
	if err != nil {
		log.Fatalln("Error while creating outputfile for image: ", err)
		return
	}
	defer outputFile.Close()

	err = jpeg.Encode(outputFile, image, nil)
	if err != nil {
		log.Fatalln("Error while saving image to disk: ", err)
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
