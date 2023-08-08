package testing

import (
	"fmt"
	"gocv.io/x/gocv"
	"image-matcher/image_matcher/service"
	"log"
)

var CommandMapping = map[string]func([]string){
	"register": registerImages,
	"compare":  compareTwoImages,
	"match":    matchToDatabase,
	"test":     test,
}

func registerImages(arguments []string) {
	if len(arguments) < 2 {
		log.Fatal("not enough arguments!")
	}

	imagePath := arguments[0]
	imageAnalyzer := arguments[1]

	images := loadImagesFromPath(imagePath)

	err := service.AnalyzeAndSave(images, imageAnalyzer)

	if err != nil {
		log.Fatal(err)
	}
}

func compareTwoImages(arguments []string) {
	if len(arguments) < 4 {
		log.Fatal("not enough arguments!")
	}

	imagePath1 := arguments[0]
	imagePath2 := arguments[1]
	imageAnalyzer := arguments[2]
	imageMatcher := arguments[3]
	debug := len(arguments) > 4 && arguments[4] == "debug"

	image1 := loadImage(imagePath1)
	image2 := loadImage(imagePath2)

	isMatch, err := service.AnalyzeAndMatchTwoImages(*image1, *image2, imageAnalyzer, imageMatcher, debug)
	if err != nil {
		log.Fatal(err)
	}

	if isMatch {
		log.Println(fmt.Sprintf("%s and %s are a match", image1.ExternalReference, image2.ExternalReference))
	} else {
		log.Println(fmt.Sprintf("%s and %s don't match", image1.ExternalReference, image2.ExternalReference))
	}
}

func matchToDatabase(arguments []string) {
	if len(arguments) < 3 {
		log.Fatal("not enough arguments!")
	}

	imagePath := arguments[0]
	imageAnanlyzer := arguments[1]
	imageMatcher := arguments[2]

	image := loadImage(imagePath)

	matchReference, err := service.MatchAgainstDatabase(*image, imageAnanlyzer, imageMatcher)
	if err != nil {
		log.Fatal(err)
	}

	if matchReference != "" {
		log.Println("image matched to ", matchReference)
	} else {
		log.Println("image did not match")
	}
}

func test(args []string) {
	path1 := args[0]
	path2 := args[1]

	println(path1, path2)

	image := gocv.IMRead(path1, gocv.IMReadGrayScale)
	image2 := gocv.IMRead(path2, gocv.IMReadGrayScale)

	orb := gocv.NewORB()
	_, desc1 := orb.DetectAndCompute(image, gocv.NewMat())  //maybe mask needed?
	_, desc2 := orb.DetectAndCompute(image2, gocv.NewMat()) //maybe mask needed?

	println(desc1.Type(), desc2.Type())
	desc1.ConvertTo(&desc1, gocv.MatTypeCV32F)
	desc2.ConvertTo(&desc2, gocv.MatTypeCV32F)
	println(desc1.Type(), desc2.Type())

	flann := gocv.NewFlannBasedMatcher()
	flann.KnnMatch(desc1, desc2, 2)
}
