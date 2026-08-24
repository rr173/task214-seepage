package series

import (
	"math"

	"task214-seepage/internal/model"
)

// Validate enforces the business rules for an uploaded sequence.
func Validate(seq *model.SensorSequence) error {
	if seq.Stage == "" {
		return model.ErrMissingStage
	}
	switch seq.SensorType {
	case model.SensorPressure:
		if seq.Unit != "Pa" && seq.Unit != "kPa" {
			return model.ErrUnitMismatch
		}
	case model.SensorFlow:
		if seq.Unit != "m3/s" && seq.Unit != "m3/d" {
			return model.ErrUnitMismatch
		}
		for _, s := range seq.Samples {
			if s.Value < 0 {
				return model.ErrNegativeFlow
			}
		}
	default:
		return model.ErrUnitMismatch
	}
	if len(seq.Samples) == 0 {
		return model.ErrMissingStage
	}
	return nil
}

// Assess classifies a (valid) sequence as valid, gap or anomaly using simple
// heuristics on time ordering and value extremes.
func Assess(seq *model.SensorSequence) model.SequenceState {
	if len(seq.Samples) < 2 {
		return model.SeqValid
	}
	prev := seq.Samples[0].T
	maxGap := 0.0
	for _, s := range seq.Samples[1:] {
		gap := s.T - prev
		if gap < 0 {
			// out-of-order samples indicate an anomaly
			return model.SeqAnomaly
		}
		if gap > maxGap {
			maxGap = gap
		}
		prev = s.T
		if math.IsNaN(s.Value) || math.IsInf(s.Value, 0) {
			return model.SeqAnomaly
		}
	}
	// a single dominant gap much larger than the median spacing hints at a gap
	if maxGap > 5*medianGap(seq.Samples) {
		return model.SeqGap
	}
	return model.SeqValid
}

func medianGap(samples []model.Sample) float64 {
	if len(samples) < 2 {
		return 0
	}
	gaps := make([]float64, 0, len(samples)-1)
	for i := 1; i < len(samples); i++ {
		gaps = append(gaps, samples[i].T-samples[i-1].T)
	}
	n := len(gaps)
	if n == 0 {
		return 0
	}
	mid := n / 2
	if n%2 == 1 {
		return gaps[mid]
	}
	return (gaps[mid-1] + gaps[mid]) / 2
}
