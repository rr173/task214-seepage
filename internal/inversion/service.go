// Package inversion assembles the simulation context from an experiment's
// sequences and drives the numerical solver to estimate boundary parameters.
// Candidate models may be solved in parallel, while repeated solves of the same
// model version are serialised.
package inversion

import (
	"sync"
	"time"

	"task214-seepage/internal/model"
	"task214-seepage/internal/series"
	"task214-seepage/internal/solver"
	"task214-seepage/internal/store"
)

// Permeability search bounds (m^2) for porous-media inversion.
const (
	kLo       = 1e-15
	kHi       = 1e-9
	fitRelTol = 1e-3
	fitAbsTol = 1e-2
)

// Service orchestrates inversion tasks.
type Service struct {
	store *store.Store
	mu    sync.Map // modelID -> *sync.Mutex (same-version serial solving)
}

// New builds an inversion service over the given store.
func New(s *store.Store) *Service { return &Service{store: s} }

func (s *Service) lockFor(id string) *sync.Mutex {
	v, _ := s.mu.LoadOrStore(id, &sync.Mutex{})
	return v.(*sync.Mutex)
}

// ModelLock returns the mutex serialising repeated solves of one model version.
// Candidate models are solved in parallel, but the same model version is never
// solved concurrently.
func (s *Service) ModelLock(id string) *sync.Mutex { return s.lockFor(id) }

func (s *Service) fluidGeom(exp *model.Experiment) solver.FluidGeom {
	return solver.FluidGeom{
		Length:          exp.Geometry.Length,
		Porosity:        exp.Geometry.Porosity,
		Viscosity:       exp.Fluid.Viscosity,
		Compressibility: exp.Fluid.Compressibility,
	}
}

// sequenceFunc builds a linear-interpolation function from a sequence's samples
// via the shared solver implementation, so boundary reconstruction during
// inversion exactly matches how the data generator replays the same samples.
func sequenceFunc(seq *model.SensorSequence) func(float64) float64 {
	ts := make([]float64, len(seq.Samples))
	vs := make([]float64, len(seq.Samples))
	for i, smp := range seq.Samples {
		ts[i] = smp.T
		vs[i] = smp.Value
	}
	f := solver.PiecewiseLinear(ts, vs)
	return func(t float64) float64 {
		if t < ts[0] || t > ts[len(ts)-1] {
			return 0
		}
		return f(t)
	}
}

// bcFromSequences derives the Dirichlet boundary conditions. The inlet pressure
// sequence defines the driving pressure; the outlet pressure sequence (if any)
// fixes the downstream pressure, otherwise it is held at zero.
func (s *Service) bcFromSequences(inlet, outlet *model.SensorSequence) solver.BoundaryCond {
	inFunc := solver.RampBC(1e5, 0.05, 0).InletPressure
	if inlet != nil {
		inFunc = sequenceFunc(inlet)
	}
	outFunc := func(t float64) float64 { return 0 }
	if outlet != nil {
		outFunc = sequenceFunc(outlet)
	}
	return solver.BoundaryCond{InletPressure: inFunc, OutletPressure: outFunc}
}

// simConfig chooses a discretisation covering the observation time window.
func (s *Service) simConfig(obs []solver.Observation) solver.SimConfig {
	ng, nt := 41, 200
	t1 := 0.3
	for _, o := range obs {
		if o.T > t1 {
			t1 = o.T * 1.05
		}
	}
	return solver.SimConfig{NG: ng, NT: nt, T0: 0, T1: t1}
}

// AssembleContext builds the solver inputs for an experiment from its sequences.
func (s *Service) AssembleContext(expID string) (solver.FluidGeom, solver.BoundaryCond, solver.SimConfig, []solver.Observation, error) {
	exp, err := s.store.GetExperiment(expID)
	if err != nil {
		return solver.FluidGeom{}, solver.BoundaryCond{}, solver.SimConfig{}, nil, err
	}
	seqs, err := s.store.ListSensorSequences(expID)
	if err != nil {
		return solver.FluidGeom{}, solver.BoundaryCond{}, solver.SimConfig{}, nil, err
	}
	inlet, outlet, obsSeqs := series.SplitBoundaryAndObservations(seqs)
	obs := series.ToObservations(obsSeqs)
	fg := s.fluidGeom(exp)
	bc := s.bcFromSequences(inlet, outlet)
	cfg := s.simConfig(obs)
	return fg, bc, cfg, obs, nil
}

// SolveModel runs the solver for a single model without persisting the result.
func (s *Service) SolveModel(m *model.BoundaryModel, fg solver.FluidGeom, bc solver.BoundaryCond, cfg solver.SimConfig, obs []solver.Observation) *model.InversionTask {
	task := model.NewInversionTask(m.ID, m.ExperimentID)
	task.State = model.InvSolving
	count := len(obs)
	energy := solver.MeanObservationEnergy(obs)
	var k, k2, sse float64
	var interval [2]float64
	switch m.Form {
	case model.ModelUniform:
		k, sse, interval, _ = solver.InvertUniform(fg, bc, cfg, obs, kLo, kHi)
		task.EstimatedLayers = []model.Layer{{Permeability: k, ThicknessFrac: 1}}
	case model.ModelLayered:
		var k1 float64
		k1, k2, sse, interval, _ = solver.InvertLayered(fg, bc, cfg, obs, kLo, kHi)
		task.EstimatedLayers = []model.Layer{{Permeability: k1, ThicknessFrac: 0.5}, {Permeability: k2, ThicknessFrac: 0.5}}
		k = k1
	default:
		task.State = model.InvFailed
		return task
	}
	task.ResidualNorm = sse
	task.NormalizedSSE = sse / (float64(maxInt(count, 1)) * energy)
	task.IdentifiabilityLo = interval[0]
	task.IdentifiabilityHi = interval[1]
	task.ObservationsUsed = count
	if solver.FitsWell(sse, count, energy, fitRelTol, fitAbsTol) {
		task.State = model.InvConverged
	} else {
		task.State = model.InvNonIdentifiable
	}
	now := time.Now()
	task.CompletedAt = &now
	return task
}

// Run assembles the context for a model, solves it and persists the task.
func (s *Service) Run(modelID string) (*model.InversionTask, error) {
	m, err := s.store.GetBoundaryModel(modelID)
	if err != nil {
		return nil, err
	}
	fg, bc, cfg, obs, err := s.AssembleContext(m.ExperimentID)
	if err != nil {
		return nil, err
	}
	// Serialise repeated solves of the same model version.
	mu := s.lockFor(modelID)
	mu.Lock()
	defer mu.Unlock()
	task := s.SolveModel(m, fg, bc, cfg, obs)
	if err := s.store.CreateInversionTask(task); err != nil {
		return nil, err
	}
	return task, nil
}

// GetTask loads a persisted inversion task by id.
func (s *Service) GetTask(id string) (*model.InversionTask, error) {
	return s.store.GetInversionTask(id)
}

// ListTasks returns all inversion tasks of an experiment.
func (s *Service) ListTasks(expID string) ([]*model.InversionTask, error) {
	return s.store.ListInversionTasks(expID)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
