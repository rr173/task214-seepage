package model

import "time"

// ReleaseVersion is an immutable published inversion result for an experiment.
// A sealed experiment can only spawn new release versions, never mutate old ones.
type ReleaseVersion struct {
	ID              string    `json:"id"`
	ExperimentID    string    `json:"experiment_id"`
	SourceModelID   string    `json:"source_model_id"`
	SourceInversionID string  `json:"source_inversion_id"`
	Permeability    float64   `json:"permeability_m2"`
	IntervalLo      float64   `json:"interval_lo_m2"`
	IntervalHi      float64   `json:"interval_hi_m2"`
	Notes           string    `json:"notes"`
	Seq             int       `json:"seq"`
	CreatedAt       time.Time `json:"created_at"`
}

// NewReleaseVersion builds a published version with the given sequence number.
func NewReleaseVersion(expID, modelID, invID string, perm, lo, hi float64, seq int, notes string) *ReleaseVersion {
	return &ReleaseVersion{
		ID:               genID("rel_"),
		ExperimentID:     expID,
		SourceModelID:    modelID,
		SourceInversionID: invID,
		Permeability:     perm,
		IntervalLo:       lo,
		IntervalHi:       hi,
		Notes:            notes,
		Seq:              seq,
		CreatedAt:        time.Now(),
	}
}
