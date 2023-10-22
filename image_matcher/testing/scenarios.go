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

	threshold, err := strconv.ParseFloat(thresholdString, 64)
	if err != nil {
		log.Println("couldn't convert threshold")
	}

	var scenarioRuntime, extractionTime, matchingTime time.Duration
	var classEval statistics.ClassificationEvaluation

	if algorithm == image_handling.PHASH {
		startTime := time.Now()
		classEval, extractionTime, matchingTime = runPHashScenario(scenario, int(threshold))
		scenarioRuntime = time.Since(startTime)
	} else {
		startTime := time.Now()
		classEval, extractionTime, matchingTime = runFeatureBasedScenario(scenario, algorithm, threshold)
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
) (statistics.ClassificationEvaluation, time.Duration, time.Duration) {
	searchImages := service.GetSearchImages(scenario)
	var totalExtractionTime, totalMatchingTime time.Duration
	imageEvaluations := make([]statistics.SearchImagePHashEval, len(*searchImages))

	classificationEval := statistics.ClassificationEvaluation{}

	for _, searchImage := range *searchImages {
		path := fmt.Sprintf("images/variations/%s/%s.png", scenario, searchImage.ExternalReference)
		rawImage := image_handling.LoadRawImage(path)
		matchedReferences, err, extractionTime, matchingTime := service.MatchImageAgainstDatabasePHash(rawImage, maxHammingDistance)
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
	searchImages = nil
	imageEvaluations = nil

	statistics.WriteOverallEvalToCSV(scenario, image_handling.PHASH, &classificationEval, totalExtractionTime, totalMatchingTime)
	statistics.WritePHashImageEvalToCSV(scenario, &imageEvaluations)

	return classificationEval, totalExtractionTime, totalMatchingTime
}

func runFeatureBasedScenario(
	scenario string,
	algorithm string,
	similarityThreshold float64,
) (*statistics.ClassificationEvaluation, time.Duration, time.Duration) {
	searchImages := service.GetSearchImages(scenario)
	var totalExtractionTime, totalMatchingTime time.Duration
	imageEvaluations := make([]statistics.SearchImageFeatureBasedEval, len(*searchImages))

	classificationEval := statistics.ClassificationEvaluation{}

	for _, searchImage := range *searchImages {
		path := fmt.Sprintf("images/variations/%s/%s.png", scenario, searchImage.ExternalReference)
		rawImage := image_handling.LoadRawImage(path)
		matchedReferences, err, searchImageDescriptors, extractionTime, matchingTime :=
			service.MatchAgainstDatabaseFeatureBased(
				rawImage,
				algorithm,
				image_handling.BRUTE_FORCE_MATCHER,
				similarityThreshold,
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
	statistics.WriteOverallEvalToCSV(scenario, algorithm, &classificationEval, totalExtractionTime, totalMatchingTime)
	statistics.WriteFeatureBasedImageEvalToCSV(scenario, algorithm, &imageEvaluations)

	searchImages = nil
	imageEvaluations = nil

	return &classificationEval, totalExtractionTime, totalMatchingTime
}
