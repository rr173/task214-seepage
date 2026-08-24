// Package service composes the domain services into a single entry point used
// by the HTTP layer.
package service

import (
	"task214-seepage/internal/boundary"
	"task214-seepage/internal/compare"
	"task214-seepage/internal/inversion"
	"task214-seepage/internal/release"
	"task214-seepage/internal/series"
	"task214-seepage/internal/store"
	"task214-seepage/internal/trial"
)

// Services aggregates every domain service behind one handle.
type Services struct {
	Store     *store.Store
	Trial     *trial.Service
	Series    *series.Service
	Boundary  *boundary.Service
	Inversion *inversion.Service
	Compare   *compare.Service
	Release   *release.Service
}

// NewServices wires the store and all domain services together.
func NewServices(s *store.Store) *Services {
	inv := inversion.New(s)
	return &Services{
		Store:     s,
		Trial:     trial.New(s),
		Series:    series.New(s),
		Boundary:  boundary.New(s),
		Inversion: inv,
		Compare:   compare.New(s, inv),
		Release:   release.New(s),
	}
}
