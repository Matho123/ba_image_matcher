package testing

import (
	"fmt"
	"image_matcher/image-handling"
	"image_matcher/service"
	"image_matcher/statistics"
	"log"
	"time"
)

const (
	IDENTICAL  = "identical"
	SCALED     = "scaled"
	ROTATED    = "rotated"
	MIRRORED   = "mirrored"
	MOVED      = "moved"
	BACKGROUND = "background"
	PART       = "part"
	MIXED      = "mixed"
)

var scenarios = []string{IDENTICAL, SCALED, ROTATED, MIRRORED, MOVED, BACKGROUND, PART, MIXED}

func runSingleScenario(
	scenario string, analyzingAlgorithm string, matchingAlgorithm string, threshold float64, debug bool,
) {
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

func runAllScenarios(analyzingAlgorithm string, matchingAlgorithm string, threshold float64) {
	for _, scenario := range scenarios {
		runSingleScenario(scenario, analyzingAlgorithm, matchingAlgorithm, threshold, false)
	}
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

	thresholdString := fmt.Sprintf("%d", maxHammingDistance)

	statistics.WriteOverallEvalToCSV(
		scenario, image_handling.PHASH, "", thresholdString, &classificationEval, totalExtractionTime,
		totalMatchingTime,
	)
	statistics.WritePHashImageEvalToCSV(scenario, thresholdString, &imageEvaluations)

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

	thresholdString := fmt.Sprintf("%.2f", similarityThreshold)

	statistics.WriteOverallEvalToCSV(
		scenario, analyzingAlgorithm, matchingAlgorithm, thresholdString, &classificationEval,
		totalExtractionTime, totalMatchingTime,
	)
	statistics.WriteFeatureBasedImageEvalToCSV(
		scenario, analyzingAlgorithm, matchingAlgorithm, thresholdString, &imageEvaluations,
	)

	searchImages = nil
	imageEvaluations = nil

	return &classificationEval, totalExtractionTime, totalMatchingTime
}
