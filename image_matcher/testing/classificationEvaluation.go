package testing

import "fmt"

type ClassificationEvaluation struct {
	TP, TN, FP, FN int
}

func (c *ClassificationEvaluation) Recall() float64 {
	return float64(c.TP) / float64(c.TP+c.FN)
}

func (c *ClassificationEvaluation) Specificity() float64 {
	return float64(c.TN) / float64(c.TN+c.FP)
}

func (c *ClassificationEvaluation) BalancedAccuracy() float64 {
	return (c.Recall() + c.Specificity()) / 2
}

func (c *ClassificationEvaluation) evaluateClassification(matchedRefs *[]string, originalRef *string) string {
	amountMatched := len(*matchedRefs)
	if *originalRef == "" && amountMatched > 0 {
		c.FP++
		return "false-positive"
	} else if *originalRef != "" && amountMatched == 0 {
		c.FN++
		return "false-negative"
	} else if *originalRef != "" && containsOriginalRef(matchedRefs, originalRef) {
		c.TP++
		return "true-positive"
	} else if *originalRef == "" && !containsOriginalRef(matchedRefs, originalRef) {
		c.TN++
		return "true-negative"
	}
	return ""
}

func (c *ClassificationEvaluation) string() string {
	return fmt.Sprintf(
		"TP: %d, FP: %d, TN: %d, FN: %d, Recall: %.2f, Specificity: %.2f, balanced-acc: %.2f",
		c.TP,
		c.FP,
		c.TN,
		c.FN,
		c.Recall(),
		c.Specificity(),
		c.BalancedAccuracy(),
	)
}

func containsOriginalRef(matchedRefs *[]string, originalRef *string) bool {
	for _, matchedRef := range *matchedRefs {
		if matchedRef == *originalRef {
			return true
		}
	}
	return false
}
