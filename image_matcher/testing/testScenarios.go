package testing

import (
	"fmt"
	"image_matcher/image_analyzer"
	"image_matcher/image_handling"
	"image_matcher/image_matching"
	"image_matcher/image_service"
	"image_matcher/statistics"
	"log"
	"time"
)

var FEATURE_BASE_THRESHOLDS = []float64{0.2, 0.3, 0.4, 0.5, 0.6, 0.7}

var PHASH_THRESHOLDS = []float64{4, 8, 12, 16, 20, 24}

func runAllForEachAlgorithm([]string) {
	//for _, threshold := range PHASH_THRESHOLDS {
	//	runAllScenarios(image_handling.PHASH, "", threshold)
	//}

	/*	for _, threshold := range FEATURE_BASE_THRESHOLDS {
			//runAllScenarios(image_handling.BRISK, image_handling.BRUTE_FORCE_MATCHER, threshold)
			runFeatureBasedScenario(image_service.MIXED, image_analyzer.BRISK, image_matching.FLANN_BASED_MATCHER, threshold, false)
		}

		for _, threshold := range FEATURE_BASE_THRESHOLDS {
			//runAllScenarios(image_handling.SIFT, image_handling.BRUTE_FORCE_MATCHER, threshold)
			runFeatureBasedScenario(image_service.MIXED, image_analyzer.SIFT, image_matching.FLANN_BASED_MATCHER, threshold, false)
		}*/

	for _, threshold := range FEATURE_BASE_THRESHOLDS {
		//runAllScenarios(image_handling.ORB, image_handling.BRUTE_FORCE_MATCHER, threshold)
		runFeatureBasedScenario(image_service.MIXED, image_analyzer.ORB, image_matching.FLANN_BASED_MATCHER, threshold, false)
	}
}

func runAllScenarios(analyzingAlgorithm string, matchingAlgorithm string, threshold float64) {
	for _, scenario := range image_service.SCENARIOS {
		runSingleScenario(scenario, analyzingAlgorithm, matchingAlgorithm, threshold, false)
	}
}

func runSingleScenario(
	scenario string, analyzingAlgorithm string, matchingAlgorithm string, threshold float64, debug bool,
) {
	var scenarioRuntime, extractionTime, matchingTime time.Duration
	var classEval *statistics.ClassificationEvaluation

	if analyzingAlgorithm == image_analyzer.PHASH {
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
	searchImages := image_service.GetSearchImages(scenario)
	var totalExtractionTime, totalMatchingTime time.Duration
	var imageEvaluations []statistics.SearchImagePHashEval

	classificationEval := statistics.ClassificationEvaluation{}

	for _, searchImage := range *searchImages {
		log.Println("Matching", searchImage.ExternalReference)

		path := fmt.Sprintf("images/variations/%s/%s.png", scenario, searchImage.ExternalReference)
		rawImage := image_handling.LoadRawImage(path)
		matchedReferences, err, extractionTime, matchingTime :=
			image_service.MatchImageAgainstDatabasePHash(rawImage, maxHammingDistance, debug)
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
		scenario, image_analyzer.PHASH, "", thresholdString, &classificationEval, totalExtractionTime,
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
	searchImages := image_service.GetSearchImages(scenario)
	var totalExtractionTime, totalMatchingTime time.Duration
	var imageEvaluations []statistics.SearchImageFeatureBasedEval

	classificationEval := statistics.ClassificationEvaluation{}

	for _, searchImage := range *searchImages {
		log.Println("Matching", searchImage.ExternalReference)

		path := fmt.Sprintf("images/variations/%s/%s.png", scenario, searchImage.ExternalReference)
		rawImage := image_handling.LoadRawImage(path)
		matchedReferences, err, searchImageDescriptors, extractionTime, matchingTime :=
			image_service.MatchAgainstDatabaseFeatureBased(
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
