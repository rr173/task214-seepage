package model

import "time"

// ExperimentState enumerates the lifecycle of a seepage experiment.
type ExperimentState string

const (
	// ExpPrepared is the initial state after creation.
	ExpPrepared ExperimentState = "prepared"
	// ExpSampling means sensor sequences are being collected.
	ExpSampling ExperimentState = "sampling"
	// ExpPendingInversion means data is ready for inversion.
	ExpPendingInversion ExperimentState = "pending_inversion"
	// ExpCompleted means at least one inversion/release has been produced.
	ExpCompleted ExperimentState = "completed"
	// ExpSealed is a terminal state; no further mutations are allowed.
	ExpSealed ExperimentState = "sealed"
)

// Geometry describes the porous column sample.
type Geometry struct {
	Length   float64 `json:"length_m"`
	Diameter float64 `json:"diameter_m"`
	Porosity float64 `json:"porosity"`
}

// CrossArea returns the column cross-sectional area in m^2.
func (g Geometry) CrossArea() float64 {
	r := g.Diameter / 2
	return 3.141592653589793 * r * r
}

// Fluid describes the saturating fluid and its compressibility.
type Fluid struct {
	Viscosity       float64 `json:"viscosity_pas"`
	Density         float64 `json:"density_kgm3"`
	Compressibility float64 `json:"compressibility_pa"`
}

// Experiment is the root aggregate of the service.
type Experiment struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Geometry  Geometry        `json:"geometry"`
	Fluid     Fluid           `json:"fluid"`
	State     ExperimentState `json:"state"`
	CreatedAt time.Time       `json:"created_at"`
	SealedAt  *time.Time      `json:"sealed_at,omitempty"`
}

var expTransitions = map[ExperimentState][]ExperimentState{
	ExpPrepared:         {ExpSampling, ExpSealed},
	ExpSampling:         {ExpPendingInversion, ExpSealed},
	ExpPendingInversion: {ExpSealed},
	ExpCompleted:        {ExpSealed},
	ExpSealed:           {},
}

// CanTransition reports whether the experiment may move to the given state.
func (e *Experiment) CanTransition(to ExperimentState) bool {
	for _, s := range expTransitions[e.State] {
		if s == to {
			return true
		}
	}
	return false
}

// NewExperiment builds a prepared experiment with a generated id.
func NewExperiment(name string, g Geometry, f Fluid) *Experiment {
	return &Experiment{
		ID:        genID("exp_"),
		Name:      name,
		Geometry:  g,
		Fluid:     f,
		State:     ExpPrepared,
		CreatedAt: time.Now(),
	}
}
