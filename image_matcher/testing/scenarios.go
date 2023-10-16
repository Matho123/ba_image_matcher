package testing

import (
	"fmt"
	"image_matcher/file-handling"
	"image_matcher/service"
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
	var classEval ClassificationEvaluation
	var imageEvals *[]file_handling.SearchImageEval

	if scenario == "pHash" {
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
	log.Println("Eval: ", classEval.string())

	file_handling.WriteOverallEvalToCSV(scenario, &classEval, extractionTime, matchingTime)
	file_handling.WriteImageEvalToCSV(scenario, imageEvals)
}

func runPHashScenario(
	scenario string,
	maxHammingDistance int,
) (*[]file_handling.SearchImageEval, ClassificationEvaluation, time.Duration, time.Duration) {
	searchImages := service.GetSearchImages(scenario)
	var totalExtractionTime, totalMatchingTime time.Duration
	var imageEvaluations []file_handling.SearchImageEval

	classificationEval := ClassificationEvaluation{0, 0, 0, 0}

	for _, searchImage := range searchImages {
		path := fmt.Sprintf("images/variations/%s/%s.png", scenario, searchImage.ExternalReference)
		rawImage := file_handling.LoadRawImage(path)
		matchedReferences, err, extractionTime, matchingTime := service.MatchImageAgainstDatabasePHash(*rawImage, maxHammingDistance)
		if err != nil {
			log.Println("error while matching", searchImage.ExternalReference, "against database!")
		}
		totalExtractionTime += extractionTime
		totalMatchingTime += matchingTime

		class := classificationEval.evaluateClassification(&matchedReferences, &searchImage.OriginalReference)
		imageEvaluations = append(
			imageEvaluations,
			file_handling.SearchImageEval{ExternalReference: searchImage.ExternalReference, ClassEval: class},
		)
	}

	return &imageEvaluations, classificationEval, totalExtractionTime, totalMatchingTime
}

func runFeatureBasedScenario(
	scenario string,
	algorithm string,
	similarityThreshold float64,
) (*[]file_handling.SearchImageEval, ClassificationEvaluation, time.Duration, time.Duration) {
	startTime := time.Now()
	return &[]file_handling.SearchImageEval{file_handling.SearchImageEval{}}, ClassificationEvaluation{}, time.Since(startTime),
		time.Since(startTime)
}
