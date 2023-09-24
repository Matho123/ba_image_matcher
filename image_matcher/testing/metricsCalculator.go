package testing

type metrics struct {
	tp, tn, fp, fn int
}

func (m metrics) calculateRecall() float64 {
	return float64(m.tp / (m.tp + m.fn))
}

func (m metrics) calculateSpecificity() float64 {
	return float64(m.tn / (m.tn + m.fp))
}

func (m metrics) calculateBalancedAccuracy() float64 {
	return (m.calculateRecall() + m.calculateSpecificity()) / 2
}
