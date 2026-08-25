// Package boundary manages the catalogue of candidate boundary models that are
// inverted against the experiment observations.
package boundary

import (
	"errors"

	"task214-seepage/internal/model"
	"task214-seepage/internal/store"
)

// Service orchestrates boundary-model operations.
type Service struct {
	store *store.Store
}

// New builds a boundary service over the given store.
func New(s *store.Store) *Service { return &Service{store: s} }

// CreateUniform registers a single-layer homogeneous candidate.
func (s *Service) CreateUniform(expID, name string, k float64) (*model.BoundaryModel, error) {
	if k <= 0 {
		return nil, errors.New("permeability must be positive")
	}
	m := model.NewUniformModel(expID, name, k)
	if err := s.store.CreateBoundaryModel(m); err != nil {
		return nil, err
	}
	return m, nil
}

// CreateLayered registers a two-layer (inlet half / outlet half) candidate.
func (s *Service) CreateLayered(expID, name string, k1, k2 float64) (*model.BoundaryModel, error) {
	if k1 <= 0 || k2 <= 0 {
		return nil, errors.New("permeability must be positive")
	}
	m := model.NewLayeredModel(expID, name, k1, k2)
	if err := s.store.CreateBoundaryModel(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Get returns a model by id.
func (s *Service) Get(id string) (*model.BoundaryModel, error) { return s.store.GetBoundaryModel(id) }

// List returns all models of an experiment.
func (s *Service) List(expID string) ([]*model.BoundaryModel, error) {
	return s.store.ListBoundaryModels(expID)
}

// SetState transitions a model between draft/runnable/retired/confirmed.
func (s *Service) SetState(id string, to model.ModelState) error {
	valid := map[model.ModelState]bool{
		model.ModelDraft:     true,
		model.ModelRunnable:  true,
		model.ModelRetired:   true,
		model.ModelConfirmed: true,
	}
	if !valid[to] {
		return model.ErrInvalidState
	}
	return s.store.UpdateBoundaryModelState(id, to)
}
