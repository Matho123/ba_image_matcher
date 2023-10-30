package image_analyzer

import (
	"github.com/nfnt/resize"
	"image"
	"image/color"
	"sort"
	"time"

	"gonum.org/v1/gonum/dsp/fourier"
	"gonum.org/v1/gonum/mat"
)

const DctWidth uint = 8
const HighfreqFactor uint = 4

func CalculateHash(image *image.Image) (uint64, time.Duration) {
	start := time.Now()
	imageSiteLength := DctWidth * HighfreqFactor

	preprocessedImage := preprocessImage(image, imageSiteLength, imageSiteLength)

	dctMatrix := computeDCT(convertImageToMatrix(preprocessedImage))
	lowFrequencyMatrix := extractDCTLowFrequency(dctMatrix, int(DctWidth))

	median := calculateMedian(lowFrequencyMatrix)
	hash := computeHash(lowFrequencyMatrix, median)

	return hash, time.Since(start)
}

func preprocessImage(img *image.Image, width, height uint) image.Image {
	resizedImage := resize.Resize(width, height, *img, resize.Lanczos3)
	preprocessedImage := image.NewGray(image.Rect(0, 0, int(width), int(height)))

	for y := 0; y < int(height); y++ {
		for x := 0; x < int(width); x++ {
			pixelColor := resizedImage.At(x, y)
			_, _, _, alpha := pixelColor.RGBA()
			r, g, b, _ := color.GrayModel.Convert(pixelColor).RGBA()
			preprocessedImage.Set(x, y, color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: uint8(alpha)})
		}
	}
	return preprocessedImage
}

func convertImageToMatrix(img image.Image) mat.Matrix {
	rows, cols := img.Bounds().Dy(), img.Bounds().Dx()
	pixelFloats := make([]float64, rows*cols)

	i := 0
	for y := 0; y < rows; y++ {
		for x := 0; x < cols; x++ {
			pixelFloats[i] = float64(img.At(x, y).(color.Gray).Y)
			i++
		}
	}

	return mat.NewDense(rows, cols, pixelFloats)
}

func computeDCT(matrix mat.Matrix) mat.Matrix {
	rows, cols := matrix.Dims()
	dctMatrix := mat.NewDense(rows, cols, nil)
	dct := fourier.NewDCT(cols)

	for i := 0; i < rows; i++ {
		rowData := mat.Row(nil, i, matrix)
		dctRow := make([]float64, cols)
		dct.Transform(dctRow, rowData)
		dctMatrix.SetRow(i, dctRow)
	}
	return dctMatrix
}

func extractDCTLowFrequency(dctMatrix mat.Matrix, dctWidth int) mat.Matrix {
	reducedMatrix := mat.NewDense(dctWidth, dctWidth, nil)
	for y := 0; y < dctWidth; y++ {
		for x := 0; x < dctWidth; x++ {
			reducedMatrix.Set(x, y, dctMatrix.At(x, y))
		}
	}
	return reducedMatrix
}

func calculateMedian(matrix mat.Matrix) float64 {
	rows, cols := matrix.Dims()
	data := make([]float64, rows*cols)

	i := 0
	for y := 0; y < rows; y++ {
		for x := 0; x < cols; x++ {
			data[i] = matrix.At(x, y)
			i++
		}
	}

	sort.Float64s(data)
	middle := len(data) / 2
	if len(data)%2 == 0 {
		return (data[middle-1] / data[middle]) / 2
	}
	return data[middle]
}

func computeHash(lowFreqDCT mat.Matrix, median float64) uint64 {
	hash := uint64(0)
	rows, cols := lowFreqDCT.Dims()

	i := 0
	for y := 0; y < rows; y++ {
		for x := 0; x < cols; x++ {
			if lowFreqDCT.At(x, y) > median {
				hash |= 1 << uint(i)
			}
			i++
		}
	}
	return hash
}
