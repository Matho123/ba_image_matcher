package testing

import (
	"fmt"
	"image_matcher/image-handling"
	"image_matcher/service"
	"image_matcher/statistics"
	"log"
	"strconv"
	"time"
)

func runScenario(args []string) {
	scenario := args[0]
	analyzingAlgorithm := args[1]
	matchingAlgorithm := args[2]
	thresholdString := args[3]
	debug := len(args) > 4 && args[4] == "debug"

	threshold, err := strconv.ParseFloat(thresholdString, 64)
	if err != nil {
		log.Println("couldn't convert threshold")
	}

	var scenarioRuntime, extractionTime, matchingTime time.Duration
	var classEval *statistics.ClassificationEvaluation

	if analyzingAlgorithm == image_handling.PHASH {
		startTime := time.Now()
		classEval, extractionTime, matchingTime = runPHashScenario(scenario, int(threshold), debug)
		scenarioRuntime = time.Since(startTime)
	} else {
		startTime := time.Now()
		classEval, extractionTime, matchingTime =
			runFeatureBasedScenario(scenario, analyzingAlgorithm, matchingAlgorithm, threshold, debug)
		scenarioRuntime = time.Since(startTime)
	}

	println("\n---------------------------------")
	println("Scenario ran for", scenarioRuntime.String())
	println("ExtractionTime", extractionTime.String())
	println("MatchingTime", matchingTime.String())
	println("Eval: ", classEval.String())
}

func runPHashScenario(
	scenario string,
	maxHammingDistance int,
	debug bool,
) (*statistics.ClassificationEvaluation, time.Duration, time.Duration) {
	searchImages := service.GetSearchImages(scenario)
	var totalExtractionTime, totalMatchingTime time.Duration
	var imageEvaluations []statistics.SearchImagePHashEval

	classificationEval := statistics.ClassificationEvaluation{}

	for _, searchImage := range *searchImages {
		log.Println("Matching", searchImage.ExternalReference)

		path := fmt.Sprintf("images/variations/%s/%s.png", scenario, searchImage.ExternalReference)
		rawImage := image_handling.LoadRawImage(path)
		matchedReferences, err, extractionTime, matchingTime :=
			service.MatchImageAgainstDatabasePHash(rawImage, maxHammingDistance, debug)
		if err != nil {
			log.Println("error while matching", searchImage.ExternalReference, "against database!")
		}
		totalExtractionTime += extractionTime
		totalMatchingTime += matchingTime

		class := classificationEval.EvaluateClassification(matchedReferences, &searchImage.OriginalReference)
		imageEvaluations = append(
			imageEvaluations,
			statistics.SearchImagePHashEval{
				ExternalReference: searchImage.ExternalReference,
				ClassEval:         class,
				ExtractionTime:    extractionTime.String(),
				MatchingTime:      matchingTime.String(),
			},
		)

		//release memory
		rawImage.Data = nil
		rawImage = nil
		matchedReferences = nil
	}

	statistics.WriteOverallEvalToCSV(scenario, image_handling.PHASH, &classificationEval, totalExtractionTime, totalMatchingTime)
	statistics.WritePHashImageEvalToCSV(scenario, &imageEvaluations)

	searchImages = nil
	imageEvaluations = nil

	return &classificationEval, totalExtractionTime, totalMatchingTime
}

func runFeatureBasedScenario(
	scenario string,
	analyzingAlgorithm string,
	matchingAlgorithm string,
	similarityThreshold float64,
	debug bool,
) (*statistics.ClassificationEvaluation, time.Duration, time.Duration) {
	searchImages := service.GetSearchImages(scenario)
	var totalExtractionTime, totalMatchingTime time.Duration
	var imageEvaluations []statistics.SearchImageFeatureBasedEval

	classificationEval := statistics.ClassificationEvaluation{}

	for _, searchImage := range *searchImages {
		log.Println("Matching", searchImage.ExternalReference)

		path := fmt.Sprintf("images/variations/%s/%s.png", scenario, searchImage.ExternalReference)
		rawImage := image_handling.LoadRawImage(path)
		matchedReferences, err, searchImageDescriptors, extractionTime, matchingTime :=
			service.MatchAgainstDatabaseFeatureBased(
				rawImage,
				analyzingAlgorithm,
				matchingAlgorithm,
				similarityThreshold,
				debug,
			)
		if err != nil {
			log.Println("error while matching", searchImage.ExternalReference, "against database!")
		}

		totalExtractionTime += extractionTime
		totalMatchingTime += matchingTime

		class := classificationEval.EvaluateClassification(matchedReferences, &searchImage.OriginalReference)
		imageEvaluations = append(
			imageEvaluations,
			statistics.SearchImageFeatureBasedEval{
				ExternalReference:   searchImage.ExternalReference,
				ClassEval:           class,
				NumberOfDescriptors: searchImageDescriptors.Rows(),
				ExtractionTime:      extractionTime.String(),
				MatchingTime:        matchingTime.String(),
			},
		)

		//release memory
		rawImage.Data = nil
		rawImage = nil
		matchedReferences = nil
		searchImageDescriptors.Close()
	}

	statistics.WriteOverallEvalToCSV(scenario, analyzingAlgorithm, &classificationEval, totalExtractionTime, totalMatchingTime)
	statistics.WriteFeatureBasedImageEvalToCSV(scenario, analyzingAlgorithm, &imageEvaluations)

	searchImages = nil
	imageEvaluations = nil

	return &classificationEval, totalExtractionTime, totalMatchingTime
}
