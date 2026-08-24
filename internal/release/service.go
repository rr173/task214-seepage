// Package release publishes immutable inversion results as versioned records.
// A sealed experiment can only spawn new release versions; old versions are
// never mutated.
package release

import (
	"task214-seepage/internal/model"
	"task214-seepage/internal/store"
)

// Service orchestrates release-version operations.
type Service struct {
	store *store.Store
}

// New builds a release service over the given store.
func New(s *store.Store) *Service { return &Service{store: s} }

// Publish records a new version derived from a converged inversion. For a sealed
// experiment this simply appends a new version; for an open experiment it also
// promotes the experiment to completed.
func (s *Service) Publish(expID, modelID, invID string, perm, lo, hi float64, notes string) (*model.ReleaseVersion, error) {
	exp, err := s.store.GetExperiment(expID)
	if err != nil {
		return nil, err
	}
	seq, err := s.store.NextReleaseSeq(expID)
	if err != nil {
		return nil, err
	}
	r := model.NewReleaseVersion(expID, modelID, invID, perm, lo, hi, seq, notes)
	if err := s.store.CreateReleaseVersion(r); err != nil {
		return nil, err
	}
	if exp.State == model.ExpPendingInversion {
		_ = s.store.UpdateExperimentState(expID, model.ExpCompleted)
	}
	return r, nil
}

// Get returns a release version by id.
func (s *Service) Get(id string) (*model.ReleaseVersion, error) { return s.store.GetReleaseVersion(id) }

// List returns all release versions of an experiment in sequence order.
func (s *Service) List(expID string) ([]*model.ReleaseVersion, error) {
	return s.store.ListReleaseVersions(expID)
}
