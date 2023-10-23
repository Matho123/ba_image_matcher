package image_handling

import (
	"github.com/disintegration/imaging"
	"image"
	"image/color"
	"image/draw"
	"math"
	"math/rand"
	"time"
)

func ResizeImage(img *image.Image, scalingFactor int) image.Image {
	newWidth := (*img).Bounds().Dx() / scalingFactor
	newHeight := (*img).Bounds().Dy() / scalingFactor

	scaled := imaging.Resize(*img, newWidth, newHeight, imaging.Lanczos)
	return scaled
}

func RotateImage(img *image.Image, angle float64) image.Image {
	croppedImage := cropImage(img)
	rotatedImage := imaging.Rotate(croppedImage, angle, color.Transparent)

	return rotatedImage
}

func MirrorImage(img *image.Image, horizontal bool) (image.Image, string) {
	var mirroredImage image.Image
	var axis string

	if horizontal {
		mirroredImage = imaging.FlipH(*img)
		axis = "Y"
	} else {
		mirroredImage = imaging.FlipV(*img)
		axis = "X"
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

func MoveMotive(img *image.Image) (image.Image, float64) {
	croppedImage := cropImage(img)
	newWidth := int(float64((*img).Bounds().Dx()) * 2)
	newHeight := int(float64((*img).Bounds().Dy()) * 2)

	newImage := imaging.New(newWidth, newHeight, color.Transparent)

	movedImage, movedDistance := pasteImageRandomly(&croppedImage, newImage)

	return movedImage, movedDistance
}

func IntegrateInOtherImage(img *image.Image) (image.Image, float64) {
	croppedImage := cropImage(img)
	biggerImage := LoadImageFromDisk("images/bigger-bg.png")

	newImage, movedDistance := pasteImageRandomly(&croppedImage, *biggerImage)

	return newImage, movedDistance
}

func pasteImageRandomly(pastedImage *image.Image, backgroundImage image.Image) (image.Image, float64) {
	pastedImagedWidth := (*pastedImage).Bounds().Dx()
	pastedImageHeight := (*pastedImage).Bounds().Dy()

	backgroundWidth := backgroundImage.Bounds().Dx()
	backgroundHeight := backgroundImage.Bounds().Dy()

	maxX := (backgroundWidth - pastedImagedWidth) / 2
	maxY := (backgroundHeight - pastedImageHeight) / 2

	rand.Seed(time.Now().UnixNano())
	movedX := rand.Intn(2*maxX+1) - maxX
	movedY := rand.Intn(2*maxY+1) - maxY

	bgImageCenterX, bgImageCenterY := backgroundWidth/2, backgroundHeight/2
	pastedImageCenterX, pastedImageCenterY := pastedImagedWidth/2, pastedImageHeight/2

	offsetX := bgImageCenterX - (pastedImageCenterX + movedX)
	offsetY := bgImageCenterY - (pastedImageCenterY + movedY)

	backgroundNRGBA := image.NewNRGBA(image.Rect(0, 0, backgroundWidth, backgroundHeight))
	draw.Draw(backgroundNRGBA, image.Rect(0, 0, backgroundWidth, backgroundHeight), backgroundImage, image.Point{}, draw.Src)

	foregroundNRGBA := image.NewNRGBA(image.Rect(0, 0, pastedImagedWidth, pastedImageHeight))
	draw.Draw(foregroundNRGBA, image.Rect(0, 0, pastedImagedWidth, pastedImageHeight), *pastedImage, image.Point{}, draw.Src)

	generatedImage := pasteImage(backgroundNRGBA, foregroundNRGBA, offsetX, offsetY)

	movedDistance := math.Sqrt(
		float64(offsetX*offsetX + offsetY*offsetY),
	)

	return generatedImage, movedDistance
}

func pasteImage(bgImage *image.NRGBA, fgImage *image.NRGBA, offsetX, offsetY int) image.Image {
	dstRect := image.Rect(offsetX, offsetY, offsetX+fgImage.Bounds().Dx(), offsetY+fgImage.Bounds().Dy())
	draw.Draw(bgImage, dstRect, fgImage, fgImage.Bounds().Min, draw.Over)

	return bgImage
}

func cropImage(img *image.Image) image.Image {
	var minX, minY = (*img).Bounds().Dx(), (*img).Bounds().Dy()
	var maxX, maxY = 0, 0

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

	croppedImage := imaging.Crop(*img, image.Rect(minX, minY, maxX, maxY))

	return croppedImage
}
