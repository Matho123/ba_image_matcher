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
	algorithm := args[0]
	scenario := args[1]
	thresholdString := args[2]

	threshold, err := strconv.ParseInt(thresholdString, 10, 0)
	if err != nil {
		log.Println("couldn't convert threshold")
	}

	var scenarioRuntime, extractionTime, matchingTime time.Duration
	var classEval statistics.ClassificationEvaluation
	var imageEvals *[]statistics.SearchImageEval

	if algorithm == image_handling.PHASH {
		startTime := time.Now()
		imageEvals, classEval, extractionTime, matchingTime = runPHashScenario(scenario, int(threshold))
		scenarioRuntime = time.Since(startTime)
	} else {
		startTime := time.Now()
		imageEvals, classEval, extractionTime, matchingTime = runFeatureBasedScenario(scenario, algorithm, float64(threshold))
		scenarioRuntime = time.Since(startTime)
	}

	log.Println("Scenario ran for", scenarioRuntime)
	log.Println("ExtractionTime", extractionTime)
	log.Println("MatchingTime", matchingTime)
	log.Println("Eval: ", classEval.String())

	statistics.WriteOverallEvalToCSV(scenario, &classEval, extractionTime, matchingTime)
	statistics.WriteImageEvalToCSV(scenario, imageEvals)
}

func runPHashScenario(
	scenario string,
	maxHammingDistance int,
) (*[]statistics.SearchImageEval, statistics.ClassificationEvaluation, time.Duration, time.Duration) {
	searchImages := service.GetSearchImages(scenario)
	var totalExtractionTime, totalMatchingTime time.Duration
	var imageEvaluations []statistics.SearchImageEval

	classificationEval := statistics.ClassificationEvaluation{}

	for _, searchImage := range searchImages {
		path := fmt.Sprintf("images/variations/%s/%s.png", scenario, searchImage.ExternalReference)
		rawImage := image_handling.LoadRawImage(path)
		matchedReferences, err, extractionTime, matchingTime := service.MatchImageAgainstDatabasePHash(*rawImage, maxHammingDistance)
		if err != nil {
			log.Println("error while matching", searchImage.ExternalReference, "against database!")
		}
		totalExtractionTime += extractionTime
		totalMatchingTime += matchingTime

		class := classificationEval.EvaluateClassification(&matchedReferences, &searchImage.OriginalReference)
		imageEvaluations = append(
			imageEvaluations,
			statistics.SearchImageEval{
				ExternalReference: searchImage.ExternalReference,
				ClassEval:         class,
				ExtractionTime:    extractionTime,
				MatchingTime:      matchingTime,
			},
		)
	}

	return &imageEvaluations, classificationEval, totalExtractionTime, totalMatchingTime
}

func runFeatureBasedScenario(
	scenario string,
	algorithm string,
	similarityThreshold float64,
) (*[]statistics.SearchImageEval, statistics.ClassificationEvaluation, time.Duration, time.Duration) {
	startTime := time.Now()
	return &[]statistics.SearchImageEval{statistics.SearchImageEval{}}, statistics.ClassificationEvaluation{}, time.Since(startTime),
		time.Since(startTime)
}
