// Package model defines the domain entities, state machines and errors for the
// porous-media seepage experiment boundary-inversion service.
package model

import "errors"

// Domain errors shared across packages.
var (
	// ErrNotFound is returned when a requested entity does not exist.
	ErrNotFound = errors.New("resource not found")
	// ErrInvalidState is returned for an illegal state transition.
	ErrInvalidState = errors.New("invalid state transition")
	// ErrSealed is returned when a mutation is attempted on a sealed experiment.
	ErrSealed = errors.New("experiment is sealed; modifications are disallowed")
	// ErrUnitMismatch is returned when a sensor unit is inconsistent with its type.
	ErrUnitMismatch = errors.New("unit is inconsistent with sensor type")
	// ErrNegativeFlow is returned when a flow sensor carries a negative value.
	ErrNegativeFlow = errors.New("negative flow value is not allowed")
	// ErrMissingStage is returned when an injection stage label is absent.
	ErrMissingStage = errors.New("time stage is required")
	// ErrDuplicateSequence is returned when an identical sequence is re-uploaded.
	ErrDuplicateSequence = errors.New("sequence already exists (idempotent)")
)
