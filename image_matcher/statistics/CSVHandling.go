package statistics

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

type SearchImageEval struct {
	ExternalReference, ClassEval string
	ExtractionTime, MatchingTime time.Duration
}

func WriteOverallEvalToCSV(
	scenario string,
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
	writeToCSV(scenario+"-overall-evaluation", &data)
}

func WriteImageEvalToCSV(scenario string, imageEvaluations *[]SearchImageEval) {
	data := [][]string{
		{"image reference", "classification", "extraction time", "matching time"},
	}
	for _, imageEvaluation := range *imageEvaluations {
		data = append(
			data,
			[]string{
				imageEvaluation.ExternalReference,
				imageEvaluation.ClassEval,
				imageEvaluation.ExtractionTime.String(),
				imageEvaluation.MatchingTime.String(),
			},
		)
	}
	writeToCSV(scenario+"-detail-evaluation", &data)
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
