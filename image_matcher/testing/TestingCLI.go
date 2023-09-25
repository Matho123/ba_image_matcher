package testing

import (
	"fmt"
	"gocv.io/x/gocv"
	"image-matcher/image_matcher/service"
	"log"
)

// SimilarityThreshold TODO: needs to be tested
const SimilarityThreshold = 0.7
const HammingDistanceThreshold = 4

var CommandMapping = map[string]func([]string){
	"register": registerImages,
	"compare":  compareTwoImages,
	"match":    matchToDatabase,
	"test":     test,
	"download": downloadOriginalImages,
	"populate": populateDatabase,
	"pop1":     pop,
}

func registerImages(arguments []string) {
	if len(arguments) < 1 {
		log.Fatal("not enough arguments!")
	}

	imagePath := arguments[0]

	images := loadImagesFromPath(imagePath)

	err := service.AnalyzeAndSaveDatabaseImage(images)
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

	isMatch, kp1, kp2, extractionTime, matchingTime, err := service.AnalyzeAndMatchTwoImages(
		*image1,
		*image2,
		imageAnalyzer,
		imageMatcher,
		SimilarityThreshold, debug)
	if err != nil {
		log.Fatal(err)
	}

	if imageAnalyzer != "phash" {
		log.Println(fmt.Sprintf(
			"Keypoints extracted: %d for image1 and %d for image2",
			len(kp1),
			len(kp2),
		))
	}

	log.Println(fmt.Sprintf(
		"Time taken for extraction: %s",
		extractionTime,
	))
	log.Println(fmt.Sprintf(
		"Time taken for matching %s",
		matchingTime,
	))

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
	imageAnalyzer := arguments[1]
	imageMatcher := arguments[2]

	image := loadImage(imagePath)

	matchReferences, err, extractionTime, matchingTime := service.MatchAgainstDatabaseFeatureBased(
		*image,
		imageAnalyzer,
		imageMatcher,
		SimilarityThreshold,
	)

	log.Println(fmt.Sprintf(
		"Time taken for extraction: %s",
		extractionTime,
	))
	log.Println(fmt.Sprintf(
		"Time taken for matching %s",
		matchingTime,
	))

	if err != nil {
		log.Fatal(err)
	}

	if len(matchReferences) > 0 {
		for _, match := range matchReferences {
			log.Println(fmt.Sprintf("image matched to %s", match))
		}
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
