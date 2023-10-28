package image_analyzer

import (
	"gocv.io/x/gocv"
	"image"
	"image/color"
	"image_matcher/image_handling"
	"time"
)

func CalculateRotationInvariantHashes(image *image.Image) ([]uint64, time.Duration) {
	orientation := getOrientation(image)
	//println(fmt.Sprintf("%.2f", orientation))

	start := time.Now()
	normalizedImage1 := image_handling.RotateImage(image, orientation)
	normalizedImage2 := image_handling.RotateImage(image, 360-(180-orientation))
	end := time.Since(start)
	//
	//image_handling.SaveImageToDisk("debug/normalized2", normalizedImage1)
	//image_handling.SaveImageToDisk("debug/normalized3", normalizedImage2)

	hash1, extractionTime1 := GetPHashValue(&normalizedImage1)
	hash2, extractionTime2 := GetPHashValue(&normalizedImage2)

	totalExtractionTime := time.Duration((extractionTime1+extractionTime2)*float64(time.Second)) + end

	return []uint64{hash1, hash2}, totalExtractionTime
}

func CalculateRotationInvariantHash(image *image.Image) (uint64, time.Duration) {
	orientation := getOrientation(image)
	//println(fmt.Sprintf("%.2f", orientation))

	start := time.Now()
	normalizedImage := image_handling.RotateImage(image, orientation)
	end := time.Since(start)

	image_handling.SaveImageToDisk("debug/normalized1", normalizedImage)

	hash, extractionTime := GetPHashValue(&normalizedImage)

	return hash, time.Duration((extractionTime)*float64(time.Second)) + end
}

func getOrientation(image *image.Image) float64 {
	grayImage, contours := getGrayImageAndContours(image)
	defer grayImage.Close()

	var biggestContour gocv.PointVector
	biggestContourArea := 0.0
	//index := 0

	for i := 0; i < contours.Size(); i++ {
		contour := contours.At(i)
		area := gocv.ContourArea(contour)
		if area > biggestContourArea {
			biggestContourArea = area
			biggestContour = contour
			//index = i
		}

	}

	//gocv.DrawContours(grayImage, *contours, index, color.RGBA{R: 255, G: 255, B: 255, A: 255}, 5)
	//gocv.IMWrite("debug/contours.png", *grayImage)

	if !biggestContour.IsNil() && len(biggestContour.ToPoints()) >= 5 {
		ellipse := gocv.FitEllipse(biggestContour)
		return ellipse.Angle
	}

	return 0
}

func getGrayImageAndContours(image *image.Image) (*gocv.Mat, *gocv.PointsVector) {
	whiteBGGrayImage :=
		image_handling.ConvertImageToGrayMatWithBackground(image, color.RGBA{R: 255, G: 255, B: 255, A: 255})

	blackBGGrayImage := image_handling.ConvertImageToGrayMatWithBackground(image, color.RGBA{A: 255})

	whiteEdges := gocv.NewMat()
	defer whiteEdges.Close()
	gocv.Canny(whiteBGGrayImage, &whiteEdges, 10, 200)

	blackEdges := gocv.NewMat()
	gocv.Canny(blackBGGrayImage, &blackEdges, 10, 200)
	defer blackEdges.Close()

	contoursBlack := gocv.FindContours(blackEdges, gocv.RetrievalExternal, gocv.ChainApproxNone)
	contoursWhite := gocv.FindContours(whiteEdges, gocv.RetrievalExternal, gocv.ChainApproxNone)
	if contoursWhite.Size() > contoursBlack.Size() {
		return &whiteBGGrayImage, &contoursWhite
	}
	return &blackBGGrayImage, &contoursBlack
}
