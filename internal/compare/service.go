// Package compare runs every candidate boundary model for an experiment in
// parallel and judges whether the observations can distinguish them
// (structural identifiability). It also summarises residual/identifiability
// statistics for reporting.
package compare

import (
	"sort"
	"sync"

	"task214-seepage/internal/inversion"
	"task214-seepage/internal/model"
	"task214-seepage/internal/solver"
	"task214-seepage/internal/store"
)

// Service orchestrates candidate-model comparison.
type Service struct {
	store *store.Store
	inv   *inversion.Service
}

// New builds a compare service.
func New(s *store.Store, inv *inversion.Service) *Service {
	return &Service{store: s, inv: inv}
}

// Result bundles the identifiability comparison with the per-model tasks.
type Result struct {
	Comparison solver.Comparison
	Tasks      []*model.InversionTask
}

// Compare solves all runnable candidate models for the experiment concurrently
// and judges structural identifiability.
func (s *Service) Compare(expID string) (*Result, error) {
	models, err := s.store.ListBoundaryModels(expID)
	if err != nil {
		return nil, err
	}
	fg, bc, cfg, obs, err := s.inv.AssembleContext(expID)
	if err != nil {
		return nil, err
	}
	if len(obs) == 0 {
		return &Result{
			Comparison: solver.Comparison{
				Identifiable: false,
				Reason:       "no internal observations available; candidate models are unconstrained and the boundary structure is non-identifiable",
			},
		}, nil
	}
	runnable := make([]*model.BoundaryModel, 0, len(models))
	formByID := map[string]model.ModelForm{}
	for _, m := range models {
		if m.State == model.ModelRunnable || m.State == model.ModelConfirmed {
			runnable = append(runnable, m)
			formByID[m.ID] = m.Form
		}
	}
	if len(runnable) == 0 {
		return nil, model.ErrNotFound
	}

	var wg sync.WaitGroup
	tasks := make([]*model.InversionTask, len(runnable))
	for i, m := range runnable {
		wg.Add(1)
		go func(i int, m *model.BoundaryModel) {
			defer wg.Done()
			// serialise repeated solves of the same model version
			mu := s.inv.ModelLock(m.ID)
			mu.Lock()
			t := s.inv.SolveModel(m, fg, bc, cfg, obs)
			_ = s.store.CreateInversionTask(t)
			mu.Unlock()
			tasks[i] = t
		}(i, m)
	}
	wg.Wait()

	sorted := make([]*model.InversionTask, len(tasks))
	copy(sorted, tasks)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].ModelID < sorted[j].ModelID })

	mr := make([]solver.ModelResult, 0, len(sorted))
	for _, t := range sorted {
		mr = append(mr, solver.ModelResult{
			Form:          string(formByID[t.ModelID]),
			SSE:           t.ResidualNorm,
			NormalizedSSE: t.NormalizedSSE,
			PrimaryK:      primaryK(t),
			Interval:      [2]float64{t.IdentifiabilityLo, t.IdentifiabilityHi},
			Converged:     t.State == model.InvConverged,
			Fits:          t.State == model.InvConverged,
		})
	}
	comp := solver.CompareModels(mr, 1e-3)
	return &Result{Comparison: comp, Tasks: sorted}, nil
}

// ResidualSummary reports per-model residual and identifiability statistics.
type ResidualSummary struct {
	ModelID           string  `json:"model_id"`
	Form              string  `json:"form"`
	State             string  `json:"state"`
	ResidualNorm      float64 `json:"residual_norm"`
	NormalizedSSE     float64 `json:"normalized_sse"`
	IdentifiabilityLo float64 `json:"identifiability_lo"`
	IdentifiabilityHi float64 `json:"identifiability_hi"`
	ObservationsUsed  int     `json:"observations_used"`
}

// Summarize compiles a residual report from completed inversion tasks.
func Summarize(tasks []*model.InversionTask, formByID map[string]string) []ResidualSummary {
	out := make([]ResidualSummary, 0, len(tasks))
	for _, t := range tasks {
		out = append(out, ResidualSummary{
			ModelID:           t.ModelID,
			Form:              formByID[t.ModelID],
			State:             string(t.State),
			ResidualNorm:      t.ResidualNorm,
			NormalizedSSE:     t.NormalizedSSE,
			IdentifiabilityLo: t.IdentifiabilityLo,
			IdentifiabilityHi: t.IdentifiabilityHi,
			ObservationsUsed:  t.ObservationsUsed,
		})
	}
	return out
}

func primaryK(t *model.InversionTask) float64 {
	if len(t.EstimatedLayers) == 0 {
		return 0
	}
	return t.EstimatedLayers[0].Permeability
}
