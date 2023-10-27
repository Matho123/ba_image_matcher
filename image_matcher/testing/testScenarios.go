package testing

import (
	"fmt"
	"image_matcher/image_analyzer"
	"image_matcher/image_handling"
	"image_matcher/image_matching"
	"image_matcher/image_service"
	"image_matcher/statistics"
	"log"
	"strconv"
	"time"
)

var featureBaseThresholds = []float64{0.2, 0.25, 0.3, 0.35, 0.4, 0.45, 0.5, 0.55, 0.6, 0.65, 0.7}

var phashThresholds = []float64{4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24}

func runAllScenariosPerAlgorithm([]string) {
	runAllScenarios(image_analyzer.PHASH, "", &phashThresholds)

	runAllScenarios(image_analyzer.SIFT, image_matching.BFMatcher, &featureBaseThresholds)
	runAllScenarios(image_analyzer.BRISK, image_matching.BFMatcher, &featureBaseThresholds)
	runAllScenarios(image_analyzer.ORB, image_matching.BFMatcher, &featureBaseThresholds)

	runFeatureBasedScenario("mixed", image_analyzer.SIFT, image_matching.FlannMatcher, &featureBaseThresholds)
	runFeatureBasedScenario("mixed", image_analyzer.BRISK, image_matching.FlannMatcher, &featureBaseThresholds)
	runFeatureBasedScenario("mixed", image_analyzer.ORB, image_matching.FlannMatcher, &featureBaseThresholds)
}

func runAllScenarios(analyzingAlgorithm string, matchingAlgorithm string, threshold *[]float64) {
	for _, scenario := range image_service.Scenarios {
		runSingleScenario(scenario, analyzingAlgorithm, matchingAlgorithm, threshold, false)
	}
}

func runSingleScenario(
	scenario string, analyzingAlgorithm string, matchingAlgorithm string, thresholds *[]float64, debug bool,
) {
	var scenarioRuntime, extractionTime, matchingTime time.Duration
	var classEvalPhash *map[int]statistics.ClassificationEvaluation
	var classEvalFeatureBased *map[float64]statistics.ClassificationEvaluation

	if analyzingAlgorithm == image_analyzer.PHASH {
		thresholdsInt := make([]int, len(*thresholds))
		for i, threshold := range *thresholds {
			thresholdsInt[i] = int(threshold)
		}
		startTime := time.Now()
		classEvalPhash, extractionTime, matchingTime = runPHashScenario(scenario, &thresholdsInt)
		scenarioRuntime = time.Since(startTime)
	} else {
		startTime := time.Now()
		classEvalFeatureBased, extractionTime, matchingTime =
			runFeatureBasedScenario(scenario, analyzingAlgorithm, matchingAlgorithm, thresholds)
		scenarioRuntime = time.Since(startTime)
	}

	println("\n---------------------------------")
	println("Scenario ran for", scenarioRuntime.String())
	println("ExtractionTime", extractionTime.String())
	println("MatchingTime", matchingTime.String())
	if analyzingAlgorithm == image_analyzer.PHASH {
		evaluation := (*classEvalPhash)[int((*thresholds)[0])]
		println("Eval: ", evaluation.String())
	} else {
		evaluation := (*classEvalFeatureBased)[(*thresholds)[0]]
		println("Eval: ", evaluation.String())
	}
}

func runPHashScenario(
	scenario string,
	thresholds *[]int,
) (*map[int]statistics.ClassificationEvaluation, time.Duration, time.Duration) {
	var totalExtractionTime, totalMatchingTime time.Duration

	classificationMap := make(map[int]statistics.ClassificationEvaluation)

	for _, threshold := range *thresholds {
		classificationMap[threshold] = statistics.ClassificationEvaluation{}
	}

	applyScenarioRun(func(searchImage image_service.SearchImageEntity, rawImage *image_handling.RawImage) {
		matchedPerThreshold, err, extractionTime, matchingTime :=
			image_service.MatchImageAgainstDatabasePHashWithMultipleThresholds(rawImage, thresholds)
		if err != nil {
			log.Println("error while matching", searchImage.ExternalReference, "against database!")
		}
		totalExtractionTime += extractionTime
		totalMatchingTime += matchingTime

		imageEvaluations := evaluateClassificationsPHash(
			&classificationMap, matchedPerThreshold, &searchImage.OriginalReference, &searchImage.ExternalReference,
			extractionTime, matchingTime,
		)
		statistics.WritePHashImageEvalToCSV(scenario, imageEvaluations)

		matchedPerThreshold = nil
	}, scenario)

	for threshold, evaluation := range classificationMap {
		statistics.WriteOverallEvalToCSV(
			scenario, image_analyzer.PHASH, "", strconv.Itoa(threshold), &evaluation, totalExtractionTime,
			totalMatchingTime,
		)
	}

	return &classificationMap, totalExtractionTime, totalMatchingTime
}

func runFeatureBasedScenario(
	scenario string,
	analyzingAlgorithm string,
	matchingAlgorithm string,
	thresholds *[]float64,
) (*map[float64]statistics.ClassificationEvaluation, time.Duration, time.Duration) {
	var totalExtractionTime, totalMatchingTime time.Duration
	classificationMap := make(map[float64]statistics.ClassificationEvaluation)

	applyScenarioRun(func(searchImage image_service.SearchImageEntity, rawImage *image_handling.RawImage) {
		matchedPerThreshold, err, searchImageDescriptors, extractionTime, matchingTime :=
			image_service.MatchAgainstDatabaseFeatureBasedWithMultipleThresholds(
				rawImage,
				analyzingAlgorithm,
				matchingAlgorithm,
				thresholds,
			)
		if err != nil {
			log.Println("error while matching", searchImage.ExternalReference, "against database!")
		}

		totalExtractionTime += extractionTime
		totalMatchingTime += matchingTime

		imageEvaluations := evaluateClassificationsFeatureBased(
			&classificationMap, matchedPerThreshold, &searchImage.OriginalReference, &searchImage.ExternalReference,
			extractionTime, matchingTime,
		)
		statistics.WriteFeatureBasedImageEvalToCSV(scenario, analyzingAlgorithm, matchingAlgorithm, imageEvaluations)

		matchedPerThreshold = nil
		searchImageDescriptors.Close()
	}, scenario)

	for threshold, evaluation := range classificationMap {
		statistics.WriteOverallEvalToCSV(
			scenario, image_analyzer.PHASH, "", fmt.Sprintf("%.2f", threshold), &evaluation,
			totalExtractionTime,
			totalMatchingTime,
		)
	}

	return &classificationMap, totalExtractionTime, totalMatchingTime
}

func applyScenarioRun(
	applyFunction func(searchImage image_service.SearchImageEntity, rawImage *image_handling.RawImage),
	scenario string,
) {
	searchImages := image_service.GetSearchImages(scenario)
	for _, searchImage := range *searchImages {
		log.Println("Matching", searchImage.ExternalReference)

		path := fmt.Sprintf("images/variations/%s/%s.png", scenario, searchImage.ExternalReference)
		rawImage := image_handling.LoadRawImage(path)

		applyFunction(searchImage, rawImage)

		//release memory
		rawImage.Data = nil
		rawImage = nil
	}
	searchImages = nil
}

func evaluateClassificationsFeatureBased(
	classificationMap *map[float64]statistics.ClassificationEvaluation, matchedMap *map[float64][]string,
	originalRef, searchImageRef *string,
	extractionTime, matchingTime time.Duration,
) *[]statistics.SearchImageFeatureBasedEval {
	var imageEvaluations []statistics.SearchImageFeatureBasedEval
	for threshold, evaluation := range *classificationMap {
		matchedRefs := (*matchedMap)[threshold]
		class := evaluation.EvaluateClassification(&matchedRefs, originalRef)
		(*classificationMap)[threshold] = evaluation

		imageEvaluations = append(
			imageEvaluations,
			statistics.SearchImageFeatureBasedEval{
				ExternalReference: *searchImageRef,
				ClassEval:         class,
				ExtractionTime:    extractionTime.String(),
				MatchingTime:      matchingTime.String(),
			},
		)
	}
	return &imageEvaluations
}

func evaluateClassificationsPHash(
	classificationMap *map[int]statistics.ClassificationEvaluation, matchedMap *map[int][]string,
	originalRef, searchImageRef *string,
	extractionTime, matchingTime time.Duration,
) *[]statistics.SearchImagePHashEval {
	var imageEvaluations []statistics.SearchImagePHashEval
	for threshold, evaluation := range *classificationMap {
		matchedRefs := (*matchedMap)[threshold]
		class := evaluation.EvaluateClassification(&matchedRefs, originalRef)
		(*classificationMap)[threshold] = evaluation

		imageEvaluations = append(
			imageEvaluations,
			statistics.SearchImagePHashEval{
				Threshold:         threshold,
				ExternalReference: *searchImageRef,
				ClassEval:         class,
				ExtractionTime:    extractionTime.String(),
				MatchingTime:      matchingTime.String(),
			},
		)
	}
	return &imageEvaluations
}
