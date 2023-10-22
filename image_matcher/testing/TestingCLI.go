package testing

import (
	"fmt"
	"gocv.io/x/gocv"
	"image_matcher/client"
	"image_matcher/image-handling"
	"image_matcher/service"
	"log"
	"time"
)

// SimilarityThreshold TODO: needs to be tested
const SimilarityThreshold = 0.4

var CommandMapping = map[string]func([]string){
	"register": registerImages,
	"compare":  compareTwoImages,
	"match":    matchToDatabase,
	"test":     test,
	"download": downloadOriginalImages,
	"populate": populateDatabase,
	"pop1":     pop,
	"scenario": runScenario,
}

func registerImages(arguments []string) {
	if len(arguments) < 1 {
		log.Fatal("not enough arguments!")
	}

	imagePath := arguments[0]

	images := image_handling.LoadImagesFromPath(imagePath)

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

	image1 := image_handling.LoadRawImage(imagePath1)
	image2 := image_handling.LoadRawImage(imagePath2)

	isMatch, kp1, kp2, extractionTime, matchingTime, err := service.AnalyzeAndMatchTwoImages(
		*image1, *image2, imageAnalyzer, imageMatcher, SimilarityThreshold, debug,
	)
	if err != nil {
		log.Fatal(err)
	}

	if imageAnalyzer != image_handling.PHASH {
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
	debug := len(arguments) > 3 && arguments[3] == "debug"

	image := image_handling.LoadRawImage(imagePath)

	var matchReferences *[]string
	var err error
	var extractionTime, matchingTime time.Duration
	if imageAnalyzer == image_handling.PHASH {
		matchReferences, err, extractionTime, matchingTime = service.MatchImageAgainstDatabasePHash(
			image,
			4,
			debug,
		)
	} else {
		matchReferences, err, _, extractionTime, matchingTime = service.MatchAgainstDatabaseFeatureBased(
			image,
			imageAnalyzer,
			imageMatcher,
			SimilarityThreshold,
			debug,
		)
	}

	if err != nil {
		log.Println(err)
	}

	println("----------------------------------------------------")
	println("Time taken for extracting Descriptors from search image:", extractionTime.String())
	println("Time taken for matching search image with all database images:", matchingTime.String())
	if len(*matchReferences) > 0 {
		println("Search image matched to database images: ")
		for _, match := range *matchReferences {
			println(match)
		}
	} else {
		println("Image did not match")
	}
}

func downloadOriginalImages([]string) {
	designs := client.GetDownloadableImageIds()

	for _, design := range designs {
		client.DownloadImageFromUrl(design.Id)
		time.Sleep(5 * time.Second)
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
