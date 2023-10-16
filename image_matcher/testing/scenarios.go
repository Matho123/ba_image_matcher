package testing

import (
	"fmt"
	"image_matcher/file-handling"
	"image_matcher/service"
	"log"
	"strconv"
	"time"
)

type SearchImageEval struct {
	externalReference, classEval string
}

func runScenario(args []string) {
	algorithm := args[0]
	scenario := args[1]
	thresholdString := args[2]

	threshold, err := strconv.ParseInt(thresholdString, 10, 0)
	if err != nil {
		log.Println("couldn't convert threshold")
	}

	var startTime time.Time
	var scenarioRuntime, extractionTime, matchingTime time.Duration
	var classEval classificationEvaluation

	if scenario == "pHash" {
		startTime = time.Now()
		_, classEval, extractionTime, matchingTime = runPHashScenario(scenario, int(threshold))
		scenarioRuntime = time.Since(startTime)
	} else {
		startTime = time.Now()
		_, classEval, extractionTime, matchingTime = runFeatureBasedScenario(scenario, algorithm, float64(threshold))
		scenarioRuntime = time.Since(startTime)
	}

	log.Println("Scenario ran for", scenarioRuntime)
	log.Println("ExtractionTime", extractionTime)
	log.Println("MatchingTime", matchingTime)
	log.Println("Eval: ", classEval.string())
}

func runPHashScenario(
	scenario string,
	maxHammingDistance int,
) ([]SearchImageEval, classificationEvaluation, time.Duration, time.Duration) {
	searchImages := service.GetSearchImages(scenario)
	var totalExtractionTime, totalMatchingTime time.Duration
	var imageEvaluations []SearchImageEval

	classificationEval := classificationEvaluation{0, 0, 0, 0}

	for _, searchImage := range searchImages {
		path := fmt.Sprintf("images/variations/%s/%s.png", scenario, searchImage.ExternalReference)
		rawImage := file_handling.LoadRawImage(path)
		matchedReferences, err, extractionTime, matchingTime := service.MatchImageAgainstDatabasePHash(*rawImage, maxHammingDistance)
		if err != nil {
			log.Println("error while matching", searchImage.ExternalReference, "against database!")
		}
		totalExtractionTime += extractionTime
		totalMatchingTime += matchingTime
		classificationEval.evaluateClassification(&matchedReferences, &searchImage.OriginalReference)
	}

	return imageEvaluations, classificationEval, totalExtractionTime, totalMatchingTime
}

func runFeatureBasedScenario(
	scenario string,
	algorithm string,
	similarityThreshold float64,
) ([]SearchImageEval, classificationEvaluation, time.Duration, time.Duration) {
	startTime := time.Now()
	return []SearchImageEval{SearchImageEval{}}, classificationEvaluation{}, time.Since(startTime), time.Since(startTime)
}
