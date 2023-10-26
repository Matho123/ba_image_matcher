package statistics

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

type SearchImagePHashEval struct {
	ExternalReference string
	ClassEval         string
	ExtractionTime    string
	MatchingTime      string
}

type SearchImageFeatureBasedEval struct {
	ExternalReference   string
	ClassEval           string
	NumberOfDescriptors int
	ExtractionTime      string
	MatchingTime        string
}

func WriteOverallEvalToCSV(
	scenario string,
	analyzer string,
	matcher string,
	threshold string,
	classEval *ClassificationEvaluation,
	extractionTime time.Duration,
	matchingTime time.Duration,
) {
	data := [][]string{
		{
			"threshold", "tp", "tn", "fp", "fn", "recall", "specificity",
			"balanced accuracy", "extraction time", "matching time",
		},
		{
			threshold,
			strconv.Itoa(classEval.TP),
			strconv.Itoa(classEval.TN),
			strconv.Itoa(classEval.FP),
			strconv.Itoa(classEval.FN),
			fmt.Sprintf("%.2f", classEval.Recall()),
			fmt.Sprintf("%.2f", classEval.Specificity()),
			fmt.Sprintf("%.2f", classEval.BalancedAccuracy()),
			extractionTime.String(),
			matchingTime.String(),
		},
	}
	appendToCSV(
		fmt.Sprintf("%s/%s-%s-overall-evaluation", scenario, analyzer, matcher),
		&data,
	)
}

func WritePHashImageEvalToCSV(scenario string, threshold string, imageEvaluations *[]SearchImagePHashEval) {
	data := [][]string{
		{"image reference", "classification", "extraction time", "matching time"},
	}
	for _, imageEvaluation := range *imageEvaluations {
		data = append(
			data,
			[]string{
				imageEvaluation.ExternalReference,
				imageEvaluation.ClassEval,
				imageEvaluation.ExtractionTime,
				imageEvaluation.MatchingTime,
			},
		)
	}
	writeNewCSV(fmt.Sprintf("%s/phash-%s-detail-evaluation", scenario, threshold), &data)
}

func WriteFeatureBasedImageEvalToCSV(
	scenario string,
	analyzer string,
	matcher string,
	threshold string,
	imageEvaluations *[]SearchImageFeatureBasedEval,
) {
	data := [][]string{
		{
			"image reference",
			"classification",
			"number of descriptors",
			"extraction time",
			"matching time",
		},
	}
	for _, imageEvaluation := range *imageEvaluations {
		data = append(
			data,
			[]string{
				imageEvaluation.ExternalReference,
				imageEvaluation.ClassEval,
				fmt.Sprintf("%d", imageEvaluation.NumberOfDescriptors),
				imageEvaluation.ExtractionTime,
				imageEvaluation.MatchingTime,
			},
		)
	}
	writeNewCSV(
		fmt.Sprintf("%s/%s-%s-%s-detail-evaluation", scenario, analyzer, matcher, threshold),
		&data,
	)
}

func writeNewCSV(fileName string, data *[][]string) {
	file, err := os.Create("test-output/csv-files/" + fileName + ".csv")
	if err != nil {
		log.Fatal("Error writing csv", err)
	}
	defer file.Close()

	csvWriter := csv.NewWriter(file)
	defer csvWriter.Flush()

	for _, row := range *data {
		err := csvWriter.Write(row)
		if err != nil {
			log.Println("Error writing row in csv: ", err)
		}
	}
}

func appendToCSV(fileName string, data *[][]string) {
	filePath := "test-output/csv-files/" + fileName + ".csv"
	_, err := os.Stat(filePath)

	fileExists := err == nil

	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Error writing csv", err)
	}
	defer file.Close()

	csvWriter := csv.NewWriter(file)
	defer csvWriter.Flush()

	for index, row := range *data {
		if index == 0 && fileExists {
			continue
		}
		err := csvWriter.Write(row)
		if err != nil {
			log.Println("Error writing row in csv: ", err)
		}
	}
}
