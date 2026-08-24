package model

import "time"

// ModelForm enumerates candidate boundary-model structures.
type ModelForm string

const (
	// ModelUniform assumes a single homogeneous permeability over the column.
	ModelUniform ModelForm = "uniform"
	// ModelLayered assumes piecewise-constant permeability per layer.
	ModelLayered ModelForm = "layered"
)

// ModelState enumerates lifecycle states of a boundary model.
type ModelState string

const (
	// ModelDraft is an unvalidated candidate.
	ModelDraft ModelState = "draft"
	// ModelRunnable has been validated and may be solved.
	ModelRunnable ModelState = "ready"
	// ModelRetired was rejected by comparison.
	ModelRetired ModelState = "retired"
	// ModelConfirmed was selected for release.
	ModelConfirmed ModelState = "confirmed"
)

// Layer is one homogeneous sub-region of a (possibly layered) medium.
type Layer struct {
	Permeability  float64 `json:"permeability_m2"`
	ThicknessFrac float64 `json:"thickness_frac"`
}

// BoundaryModel is a candidate forward model used during inversion.
type BoundaryModel struct {
	ID           string     `json:"id"`
	ExperimentID string     `json:"experiment_id"`
	Name         string     `json:"name"`
	Form         ModelForm  `json:"form"`
	Layers       []Layer    `json:"layers"`
	State        ModelState `json:"state"`
	CreatedAt    time.Time  `json:"created_at"`
}

// NewUniformModel builds a single-layer uniform candidate.
func NewUniformModel(expID, name string, k float64) *BoundaryModel {
	return &BoundaryModel{
		ID:           genID("mdl_"),
		ExperimentID: expID,
		Name:         name,
		Form:         ModelUniform,
		Layers:       []Layer{{Permeability: k, ThicknessFrac: 1}},
		State:        ModelDraft,
		CreatedAt:    time.Now(),
	}
}

// NewLayeredModel builds a two-layer candidate (inlet half, outlet half).
func NewLayeredModel(expID, name string, k1, k2 float64) *BoundaryModel {
	return &BoundaryModel{
		ID:           genID("mdl_"),
		ExperimentID: expID,
		Name:         name,
		Form:         ModelLayered,
		Layers: []Layer{
			{Permeability: k1, ThicknessFrac: 0.5},
			{Permeability: k2, ThicknessFrac: 0.5},
		},
		State:     ModelDraft,
		CreatedAt: time.Now(),
	}
}
