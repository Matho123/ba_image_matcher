package testing

import "fmt"

type classificationEvaluation struct {
	tp, tn, fp, fn int
}

func (c *classificationEvaluation) recall() float64 {
	return float64(c.tp) / float64(c.tp+c.fn)
}

func (c *classificationEvaluation) specificity() float64 {
	return float64(c.tn) / float64(c.tn+c.fp)
}

func (c *classificationEvaluation) balancedAccuracy() float64 {
	return (c.recall() + c.specificity()) / 2
}

func (c *classificationEvaluation) evaluateClassification(matchedRefs *[]string, originalRef *string) string {
	amountMatched := len(*matchedRefs)
	if *originalRef == "" && amountMatched > 0 {
		c.fp++
		return "false-positive"
	} else if *originalRef != "" && amountMatched == 0 {
		c.fn++
		return "false-negative"
	} else if *originalRef != "" && containsOriginalRef(matchedRefs, originalRef) {
		c.tp++
		return "true-positive"
	} else if *originalRef == "" && !containsOriginalRef(matchedRefs, originalRef) {
		c.tn++
		return "true-negative"
	}
	return ""
}

func (c *classificationEvaluation) string() string {
	return fmt.Sprintf(
		"tp: %d, fp: %d, tn: %d, fn: %d, recall: %.2f, specificity: %.2f, balanced-acc: %.2f",
		c.tp,
		c.fp,
		c.tn,
		c.fn,
		c.recall(),
		c.specificity(),
		c.balancedAccuracy(),
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
