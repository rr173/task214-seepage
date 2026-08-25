package model

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
	"time"
)

// SensorType classifies a recorded signal.
type SensorType string

const (
	// SensorPressure records pore pressure (Pa).
	SensorPressure SensorType = "pressure"
	// SensorFlow records volumetric flow rate (m^3/s).
	SensorFlow SensorType = "flow"
)

// SequenceState enumerates validation states of an uploaded sequence.
type SequenceState string

const (
	// SeqPending means the sequence has not been validated yet.
	SeqPending SequenceState = "pending_validation"
	// SeqValid means the sequence passed validation.
	SeqValid SequenceState = "valid"
	// SeqGap means the sequence has missing time windows.
	SeqGap SequenceState = "gap"
	// SeqAnomaly means the sequence carries anomalous values.
	SeqAnomaly SequenceState = "anomaly"
)

// Sample is a single (time, value) measurement.
type Sample struct {
	T     float64 `json:"t"`
	Value float64 `json:"value"`
}

// SensorSequence is an uploaded pressure/flow record at a column location.
type SensorSequence struct {
	ID           string        `json:"id"`
	ExperimentID string        `json:"experiment_id"`
	SensorType   SensorType    `json:"sensor_type"`
	Location     string        `json:"location"`
	XFrac        float64       `json:"x_frac"`
	Stage        string        `json:"stage"`
	Unit         string        `json:"unit"`
	Samples      []Sample      `json:"samples"`
	WindowHash   string        `json:"window_hash"`
	State        SequenceState `json:"state"`
	CreatedAt    time.Time     `json:"created_at"`
}

// ComputeWindowHash returns a stable hash over (experiment, type, location,
// stage, sorted samples). The samples are normalised to ascending time order so
// that merely reordering the same input samples leaves the hash unchanged; it is
// used for idempotent re-uploads.
func (s *SensorSequence) ComputeWindowHash() string {
	type key struct {
		EID  string
		Typ  string
		Loc  string
		Stg  string
		Samp string
	}
	samples := make([]Sample, len(s.Samples))
	copy(samples, s.Samples)
	sort.Slice(samples, func(i, j int) bool {
		if samples[i].T == samples[j].T {
			return samples[i].Value < samples[j].Value
		}
		return samples[i].T < samples[j].T
	})
	raw, _ := json.Marshal(samples)
	k := key{EID: s.ExperimentID, Typ: string(s.SensorType), Loc: s.Location, Stg: s.Stage, Samp: string(raw)}
	sum := sha256.Sum256([]byte(jsonMarshal(k)))
	return hex.EncodeToString(sum[:])
}

func jsonMarshal(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}

// NewSensorSequence builds a sequence, derives its window hash and id.
func NewSensorSequence(expID string, typ SensorType, loc string, xf float64, stage, unit string, samples []Sample) *SensorSequence {
	s := &SensorSequence{
		ID:           genID("seq_"),
		ExperimentID: expID,
		SensorType:   typ,
		Location:     loc,
		XFrac:        xf,
		Stage:        stage,
		Unit:         unit,
		Samples:      samples,
		State:        SeqPending,
		CreatedAt:    time.Now(),
	}
	s.WindowHash = s.ComputeWindowHash()
	return s
}
