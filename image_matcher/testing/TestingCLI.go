package testing

import (
	"fmt"
	"gocv.io/x/gocv"
	"image_matcher/image_analyzer"
	"image_matcher/image_handling"
	"image_matcher/image_service"
	"log"
	"strconv"
	"time"
)

const SimilarityThreshold = 0.4

var CommandMapping = map[string]func([]string){
	"register":  registerImages,
	"compare":   compareTwoImages,
	"match":     matchToDatabase,
	"scenario":  runScenario,
	"duplicate": duplicate,
	"uniques":   uniques,
	"runAll":    runAllScenariosPerAlgorithm,
	"update":    updateDatabaseWithNewHash,
}

func duplicate(arguments []string) {
	if len(arguments) < 1 {
		log.Fatal("Need a directory of images!")
	}
	populateDatabase(arguments[0])
}

func uniques(arguments []string) {
	if len(arguments) < 1 {
		log.Fatal("Need a directory of images!")
	}
	generateUniques(arguments[0])
}

func registerImages(arguments []string) {
	if len(arguments) < 1 {
		log.Fatal("not enough arguments!")
	}

	imagePath := arguments[0]

	images := image_handling.LoadImagesFromPath(imagePath)
	err := image_service.AnalyzeAndSaveDatabaseImage(images)
	if err != nil {
		log.Fatal(err)
	}
}

func compareTwoImages(arguments []string) {
	if len(arguments) < 3 {
		log.Fatal("not enough arguments!")
	}

	imagePath1 := arguments[0]
	imagePath2 := arguments[1]
	imageAnalyzer := arguments[2]

	image1 := image_handling.LoadRawImage(imagePath1)
	if image1 == nil {
		log.Fatal("Couldn't load image: ", imagePath1)
	}
	image2 := image_handling.LoadRawImage(imagePath2)
	if image2 == nil {
		log.Fatal("Couldn't load image: ", imagePath2)
	}

	var isMatch bool
	var extractionTime, matchingTime time.Duration

	if imageAnalyzer == image_analyzer.PHASH || imageAnalyzer == image_analyzer.NewAnalyzer {
		threshold := 4
		var err error
		if len(arguments) > 3 {
			threshold, err = strconv.Atoi(arguments[3])
			if err != nil || threshold < 0 {
				log.Fatal("invalid threshold value", err)
			}
		}
		isMatch, extractionTime, matchingTime =
			image_service.AnalyzeAndMatchTwoImagesHash(*image1, *image2, imageAnalyzer, threshold)
	} else {
		if len(arguments) < 4 {
			log.Fatal("not enough arguments!")
		}
		imageMatcher := arguments[3]
		threshold := 0.4
		var err error
		if len(arguments) > 4 {
			threshold, err = strconv.ParseFloat(arguments[4], 64)
			if err != nil || threshold < 0 || threshold > 1 {
				log.Fatal("invalid threshold value", err)
			}
		}

		var kp1, kp2 []gocv.KeyPoint
		isMatch, kp1, kp2, extractionTime, matchingTime, err = image_service.AnalyzeAndMatchTwoImagesFeatureBased(
			*image1, *image2, imageAnalyzer, imageMatcher, threshold, false,
		)
		log.Println(fmt.Sprintf("Keypoints extracted: %d for image1 and %d forimage2", len(kp1), len(kp2)))
		if err != nil {
			log.Fatal(err)
		}
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
	if len(arguments) < 2 {
		log.Fatal("not enough arguments!")
	}

	imagePath := arguments[0]
	imageAnalyzer := arguments[1]
	image := image_handling.LoadRawImage(imagePath)
	if image == nil {
		log.Fatal("Couldn't load image: ", imagePath)
	}

	var matchReferences *[]string
	var err error
	var extractionTime, matchingTime time.Duration
	if imageAnalyzer == image_analyzer.PHASH {
		threshold := 4
		if len(arguments) > 2 {
			threshold, err = strconv.Atoi(arguments[2])
			if err != nil || threshold < 0 {
				log.Fatal("invalid threshold value", err)
			}
		}
		matchReferences, err, extractionTime, matchingTime = image_service.MatchImageAgainstDatabasePHash(
			image,
			threshold,
			true,
		)
	} else if imageAnalyzer == image_analyzer.NewAnalyzer {
		matchReferences, _, err, extractionTime, matchingTime =
			image_service.MatchImageAgainstDatabaseHybrid(image, true)
	} else {
		if len(arguments) < 3 {
			log.Fatal("not enough arguments!")
		}
		imageMatcher := arguments[2]

		threshold := 0.4
		if len(arguments) > 3 {
			threshold, err = strconv.ParseFloat(arguments[3], 64)
			if err != nil || threshold < 0 || threshold > 1 {
				log.Fatal("invalid threshold value", err)
			}
		}
		matchReferences, err, _, extractionTime, matchingTime = image_service.MatchAgainstDatabaseFeatureBased(
			image,
			imageAnalyzer,
			imageMatcher,
			SimilarityThreshold,
			true,
		)
	}

	if err != nil {
		log.Println(err)
	}

	if matchReferences == nil {
		return
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

func runScenario(arguments []string) {
	scenario := arguments[0]
	analyzingAlgorithm := arguments[1]
	var thresholdString string
	var matchingAlgorithm string
	if analyzingAlgorithm == image_analyzer.PHASH || analyzingAlgorithm == image_analyzer.NewAnalyzer {
		if len(arguments) < 3 {
			log.Fatal("not enough arguments!")
		}
		thresholdString = arguments[2]
	} else {
		if len(arguments) < 4 {
			log.Fatal("not enough arguments!")
		}
		matchingAlgorithm = arguments[2]
		thresholdString = arguments[3]
	}

	threshold, err := strconv.ParseFloat(thresholdString, 64)
	if err != nil {
		log.Println("threshold is not valid")
	}

	if scenario == "all" {
		runAllScenarios(analyzingAlgorithm, matchingAlgorithm, &[]float64{threshold})
	} else {
		runSingleScenario(scenario, analyzingAlgorithm, matchingAlgorithm, &[]float64{threshold}, false)
	}
}
