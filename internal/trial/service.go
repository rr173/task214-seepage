// Package trial manages the experiment lifecycle: creation, validation and the
// state machine that moves a sample from preparation through inversion to a
// sealed, immutable record.
package trial

import (
	"errors"
	"time"

	"task214-seepage/internal/model"
	"task214-seepage/internal/store"
)

// Service orchestrates experiment operations.
type Service struct {
	store *store.Store
}

// New builds a trial service over the given store.
func New(s *store.Store) *Service { return &Service{store: s} }

// Create registers a new prepared experiment after validating its geometry and
// fluid properties.
func (s *Service) Create(name string, g model.Geometry, f model.Fluid) (*model.Experiment, error) {
	if name == "" {
		return nil, errors.New("experiment name is required")
	}
	if err := ValidateGeometry(g); err != nil {
		return nil, err
	}
	if err := ValidateFluid(f); err != nil {
		return nil, err
	}
	e := model.NewExperiment(name, g, f)
	if err := s.store.CreateExperiment(e); err != nil {
		return nil, err
	}
	return e, nil
}

// Get returns an experiment by id.
func (s *Service) Get(id string) (*model.Experiment, error) { return s.store.GetExperiment(id) }

// List returns all experiments.
func (s *Service) List() ([]*model.Experiment, error) { return s.store.ListExperiments() }

// Transition moves the experiment to a new state if the transition is legal and
// the experiment is not sealed.
func (s *Service) Transition(id string, to model.ExperimentState) error {
	e, err := s.store.GetExperiment(id)
	if err != nil {
		return err
	}
	if e.State == model.ExpSealed {
		return model.ErrSealed
	}
	if !e.CanTransition(to) {
		return model.ErrInvalidState
	}
	return s.store.UpdateExperimentState(id, to)
}

// Seal marks the experiment sealed and immutable, persisting the terminal
// sealed state and recording the seal timestamp.
func (s *Service) Seal(id string) error {
	e, err := s.store.GetExperiment(id)
	if err != nil {
		return err
	}
	if e.State == model.ExpSealed {
		return model.ErrSealed
	}
	return s.store.SealExperiment(id, time.Now())
}
