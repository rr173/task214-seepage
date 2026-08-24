package httpapi

import (
	"net/http"

	"task214-seepage/internal/model"
)

type createExperimentReq struct {
	Name            string  `json:"name"`
	Length          float64 `json:"length_m"`
	Diameter        float64 `json:"diameter_m"`
	Porosity        float64 `json:"porosity"`
	Viscosity       float64 `json:"viscosity_pas"`
	Density         float64 `json:"density_kgm3"`
	Compressibility float64 `json:"compressibility_pa"`
}

func (h *Server) createExperiment(w http.ResponseWriter, r *http.Request) {
	var req createExperimentReq
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body: "+err.Error())
		return
	}
	e, err := h.svc.Trial.Create(req.Name,
		model.Geometry{Length: req.Length, Diameter: req.Diameter, Porosity: req.Porosity},
		model.Fluid{Viscosity: req.Viscosity, Density: req.Density, Compressibility: req.Compressibility})
	if err != nil {
		writeError(w, statusForError(err), err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, e)
}

func (h *Server) listExperiments(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.Trial.List()
	if err != nil {
		writeError(w, statusForError(err), err.Error())
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (h *Server) getExperiment(w http.ResponseWriter, r *http.Request) {
	e, err := h.svc.Trial.Get(r.PathValue("id"))
	if err != nil {
		writeError(w, statusForError(err), err.Error())
		return
	}
	writeJSON(w, http.StatusOK, e)
}

func (h *Server) sealExperiment(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.Trial.Seal(r.PathValue("id")); err != nil {
		writeError(w, statusForError(err), err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "completed", "id": r.PathValue("id")})
}

type transitionReq struct {
	To string `json:"to"`
}

func (h *Server) transitionExperiment(w http.ResponseWriter, r *http.Request) {
	var req transitionReq
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body: "+err.Error())
		return
	}
	if err := h.svc.Trial.Transition(r.PathValue("id"), model.ExperimentState(req.To)); err != nil {
		writeError(w, statusForError(err), err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "transitioned", "id": r.PathValue("id"), "to": req.To})
}

func (h *Server) getGeometry(w http.ResponseWriter, r *http.Request) {
	e, err := h.svc.Trial.Get(r.PathValue("id"))
	if err != nil {
		writeError(w, statusForError(err), err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"geometry":      e.Geometry,
		"fluid":         e.Fluid,
		"cross_area_m2": e.Geometry.CrossArea(),
	})
}
