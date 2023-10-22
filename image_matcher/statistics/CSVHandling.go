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
	algorithm string,
	classEval *ClassificationEvaluation,
	extractionTime time.Duration,
	matchingTime time.Duration,
) {
	data := [][]string{
		{"tp", "tn", "fp", "fn", "recall", "specificity", "balanced accuracy", "extraction time", "matching time"},
		{
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
	writeToCSV(scenario+"/"+algorithm+"-overall-evaluation", &data)
}

func WritePHashImageEvalToCSV(scenario string, imageEvaluations *[]SearchImagePHashEval) {
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
	writeToCSV(scenario+"/phash-detail-evaluation", &data)
}

func WriteFeatureBasedImageEvalToCSV(
	scenario string,
	algorithm string,
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
	writeToCSV(scenario+"/"+algorithm+"-detail-evaluation", &data)
}

func writeToCSV(fileName string, data *[][]string) {
	file, err := os.Create("test-output/" + fileName + ".csv")
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
