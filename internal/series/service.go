// Package series manages uploaded pressure/flow sensor sequences, enforcing
// unit consistency, gap/anomaly detection and idempotent re-uploads.
package series

import (
	"sort"

	"task214-seepage/internal/model"
	"task214-seepage/internal/solver"
	"task214-seepage/internal/store"
)

// Service orchestrates sensor-sequence operations.
type Service struct {
	store *store.Store
}

// New builds a series service over the given store.
func New(s *store.Store) *Service { return &Service{store: s} }

// Upload validates and stores a sequence. Re-uploading an identical window is
// idempotent and returns the existing record with model.ErrDuplicateSequence.
// Sealed experiments reject new sequences.
func (s *Service) Upload(expID, typ, location string, xf float64, stage, unit string, samples []model.Sample) (*model.SensorSequence, error) {
	exp, err := s.store.GetExperiment(expID)
	if err != nil {
		return nil, err
	}
	if exp.State == model.ExpSealed {
		return nil, model.ErrSealed
	}
	if typ != string(model.SensorPressure) && typ != string(model.SensorFlow) {
		return nil, model.ErrUnitMismatch
	}
	seq := model.NewSensorSequence("", model.SensorType(typ), location, xf, stage, unit, samples)
	if err := Validate(seq); err != nil {
		return nil, err
	}
	seq.State = Assess(seq)
	stored, err := s.store.CreateSensorSequence(seq)
	if err != nil {
		return stored, err
	}
	return stored, nil
}

// Get returns a sequence by id.
func (s *Service) Get(id string) (*model.SensorSequence, error) { return s.store.GetSensorSequence(id) }

// List returns all sequences of an experiment.
func (s *Service) List(expID string) ([]*model.SensorSequence, error) {
	return s.store.ListSensorSequences(expID)
}

// Revalidate re-runs validation and updates the sequence state.
func (s *Service) Revalidate(id string) (*model.SensorSequence, error) {
	seq, err := s.store.GetSensorSequence(id)
	if err != nil {
		return nil, err
	}
	if err := Validate(seq); err != nil {
		return nil, err
	}
	seq.State = Assess(seq)
	if err := s.store.UpdateSensorState(id, seq.State); err != nil {
		return nil, err
	}
	return seq, nil
}

// SplitBoundaryAndObservations partitions sequences into Dirichlet boundary
// conditions (inlet/outlet pressure) and internal observations used for fitting.
func SplitBoundaryAndObservations(seqs []*model.SensorSequence) (inlet, outlet *model.SensorSequence, obs []*model.SensorSequence) {
	for _, seq := range seqs {
		switch seq.Location {
		case "inlet":
			inlet = seq
		case "outlet":
			outlet = seq
		default:
			obs = append(obs, seq)
		}
	}
	return
}

// ToObservations flattens internal sequences into solver observations.
func ToObservations(seqs []*model.SensorSequence) []solver.Observation {
	out := make([]solver.Observation, 0, len(seqs)*4)
	for _, seq := range seqs {
		for _, smp := range seq.Samples {
			out = append(out, solver.Observation{XFrac: seq.XFrac, T: smp.T, Value: smp.Value})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].T == out[j].T {
			return out[i].XFrac < out[j].XFrac
		}
		return out[i].T < out[j].T
	})
	return out
}
