// Package httpapi exposes the seepage inversion service over HTTP. All routes are
// rooted at /api and return JSON.
package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"task214-seepage/internal/model"
	"task214-seepage/internal/service"
)

// Server exposes the HTTP API.
type Server struct {
	svc *service.Services
}

// New builds the API server.
func New(svc *service.Services) *Server { return &Server{svc: svc} }

// Routes registers every endpoint on a ServeMux and returns it.
func (h *Server) Routes() *http.ServeMux {
	mux := http.NewServeMux()
	// experiments
	mux.HandleFunc("POST /api/experiments", h.createExperiment)
	mux.HandleFunc("GET /api/experiments", h.listExperiments)
	mux.HandleFunc("GET /api/experiments/{id}", h.getExperiment)
	mux.HandleFunc("PATCH /api/experiments/{id}/seal", h.sealExperiment)
	mux.HandleFunc("PATCH /api/experiments/{id}/transition", h.transitionExperiment)
	mux.HandleFunc("GET /api/experiments/{id}/geometry", h.getGeometry)
	// sensor sequences
	mux.HandleFunc("POST /api/experiments/{id}/sequences", h.uploadSequence)
	mux.HandleFunc("GET /api/experiments/{id}/sequences", h.listSequences)
	mux.HandleFunc("GET /api/sequences/{id}", h.getSequence)
	mux.HandleFunc("POST /api/sequences/{id}/revalidate", h.revalidateSequence)
	mux.HandleFunc("GET /api/experiments/{id}/sequences/stats", h.sequenceStats)
	// boundary models
	mux.HandleFunc("POST /api/experiments/{id}/models", h.createModel)
	mux.HandleFunc("GET /api/experiments/{id}/models", h.listModels)
	mux.HandleFunc("GET /api/models/{id}", h.getModel)
	mux.HandleFunc("PATCH /api/models/{id}/status", h.setModelStatus)
	// inversions
	mux.HandleFunc("POST /api/models/{id}/inversions", h.runInversion)
	mux.HandleFunc("GET /api/inversions/{id}", h.getInversion)
	mux.HandleFunc("GET /api/experiments/{id}/inversions", h.listInversions)
	mux.HandleFunc("POST /api/inversions/{id}/rerun", h.rerunInversion)
	// residuals / comparison
	mux.HandleFunc("GET /api/experiments/{id}/residuals/compare", h.compareModels)
	mux.HandleFunc("GET /api/inversions/{id}/identifiability", h.inversionIdentifiability)
	// release versions
	mux.HandleFunc("POST /api/experiments/{id}/releases", h.publishRelease)
	mux.HandleFunc("GET /api/experiments/{id}/releases", h.listReleases)
	mux.HandleFunc("GET /api/releases/{id}", h.getRelease)
	// self-check
	mux.HandleFunc("GET /api/selfcheck", h.selfCheck)
	return mux
}

func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func readJSON(r *http.Request, v interface{}) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}

type errorResponse struct {
	Error string `json:"error"`
}

func writeError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, errorResponse{Error: msg})
}

// statusForError maps domain errors to HTTP status codes.
func statusForError(err error) int {
	switch {
	case errors.Is(err, model.ErrSealed),
		errors.Is(err, model.ErrInvalidState),
		errors.Is(err, model.ErrUnitMismatch),
		errors.Is(err, model.ErrNegativeFlow),
		errors.Is(err, model.ErrMissingStage),
		errors.Is(err, model.ErrDuplicateSequence):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
