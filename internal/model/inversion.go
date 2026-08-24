package model

import "time"

// InversionState enumerates the lifecycle of an inversion task.
type InversionState string

const (
	// InvQueued means the task is waiting to be solved.
	InvQueued InversionState = "queued"
	// InvSolving means the solver is running.
	InvSolving InversionState = "solving"
	// InvConverged means the solver found a well-fitting estimate.
	InvConverged InversionState = "converged"
	// InvNonIdentifiable means several candidates fit equally well.
	InvNonIdentifiable InversionState = "nonidentifiable"
	// InvFailed means the solver aborted.
	InvFailed InversionState = "failed"
)

// InversionTask records the result of estimating boundary parameters.
type InversionTask struct {
	ID                string         `json:"id"`
	ModelID           string         `json:"model_id"`
	ExperimentID      string         `json:"experiment_id"`
	State             InversionState `json:"state"`
	EstimatedLayers   []Layer        `json:"estimated_layers"`
	ResidualNorm      float64        `json:"residual_norm"`
	NormalizedSSE     float64        `json:"normalized_sse"`
	IdentifiabilityLo float64        `json:"identifiability_lo"`
	IdentifiabilityHi float64        `json:"identifiability_hi"`
	ObservationsUsed  int            `json:"observations_used"`
	CreatedAt         time.Time      `json:"created_at"`
	CompletedAt       *time.Time     `json:"completed_at,omitempty"`
}

// NewInversionTask builds a queued inversion task bound to a model.
func NewInversionTask(modelID, expID string) *InversionTask {
	return &InversionTask{
		ID:           genID("inv_"),
		ModelID:      modelID,
		ExperimentID: expID,
		State:        InvQueued,
		CreatedAt:    time.Now(),
	}
}
