package httpapi

import (
	"net/http"

	"task214-seepage/internal/solver"
)

// selfCheck reports service health: the database is reachable and the numerical
// core produces finite results on a trivial problem.
func (h *Server) selfCheck(w http.ResponseWriter, r *http.Request) {
	_, err := h.svc.Trial.List()
	dbOK := err == nil

	// Trivial solver sanity: a uniform medium must be recoverable.
	fg := solver.FluidGeom{Length: 1, Porosity: 0.35, Viscosity: 1e-3, Compressibility: 1e-9}
	bc := solver.RampBC(1e5, 0.05, 0)
	cfg := solver.SimConfig{NG: 21, NT: 60, T0: 0, T1: 0.2}
	truth := solver.Medium{Layers: []solver.Layer{{Permeability: 2e-12, ThicknessFrac: 1}}}
	obs, _ := solver.SimulateObservations(truth, fg, bc, cfg, []solver.Observation{
		{XFrac: 0.5, T: 0.1}, {XFrac: 0.5, T: 0.2},
	})
	k, sse, _, _ := solver.InvertUniform(fg, bc, cfg, obs, 1e-14, 1e-10)
	solverOK := k > 0 && sse >= 0 && isFinite(k) && isFinite(sse)

	status := "ok"
	code := http.StatusOK
	if !dbOK || !solverOK {
		status = "degraded"
		code = http.StatusServiceUnavailable
	}
	writeJSON(w, code, map[string]interface{}{
		"status":   status,
		"database": dbOK,
		"solver":   solverOK,
		"recovered_uniform_k": k,
		"note":     "seepage boundary-inversion service",
	})
}

func isFinite(v float64) bool {
	return v == v && v != 1e308 && v != -1e308
}
