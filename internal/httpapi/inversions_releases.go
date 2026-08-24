package httpapi

import (
	"net/http"
)

func (h *Server) runInversion(w http.ResponseWriter, r *http.Request) {
	task, err := h.svc.Inversion.Run(r.PathValue("id"))
	if err != nil {
		writeError(w, statusForError(err), err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, task)
}

func (h *Server) getInversion(w http.ResponseWriter, r *http.Request) {
	task, err := h.svc.Inversion.GetTask(r.PathValue("id"))
	if err != nil {
		writeError(w, statusForError(err), err.Error())
		return
	}
	writeJSON(w, http.StatusOK, task)
}

func (h *Server) listInversions(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.Inversion.ListTasks(r.PathValue("id"))
	if err != nil {
		writeError(w, statusForError(err), err.Error())
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (h *Server) rerunInversion(w http.ResponseWriter, r *http.Request) {
	task, err := h.svc.Inversion.GetTask(r.PathValue("id"))
	if err != nil {
		writeError(w, statusForError(err), err.Error())
		return
	}
	newTask, err := h.svc.Inversion.Run(task.ModelID)
	if err != nil {
		writeError(w, statusForError(err), err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, newTask)
}

func (h *Server) compareModels(w http.ResponseWriter, r *http.Request) {
	res, err := h.svc.Compare.Compare(r.PathValue("id"))
	if err != nil {
		writeError(w, statusForError(err), err.Error())
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (h *Server) inversionIdentifiability(w http.ResponseWriter, r *http.Request) {
	task, err := h.svc.Inversion.GetTask(r.PathValue("id"))
	if err != nil {
		writeError(w, statusForError(err), err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"inversion_id":      task.ID,
		"state":             string(task.State),
		"identifiable_lo":   task.IdentifiabilityLo,
		"identifiable_hi":   task.IdentifiabilityHi,
		"normalized_sse":    task.NormalizedSSE,
		"residual_norm":     task.ResidualNorm,
		"observations_used": task.ObservationsUsed,
	})
}

type publishReleaseReq struct {
	ModelID       string  `json:"model_id"`
	InversionID   string  `json:"inversion_id"`
	Permeability  float64 `json:"permeability"`
	IntervalLo    float64 `json:"interval_lo"`
	IntervalHi    float64 `json:"interval_hi"`
	Notes         string  `json:"notes"`
}

func (h *Server) publishRelease(w http.ResponseWriter, r *http.Request) {
	var req publishReleaseReq
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body: "+err.Error())
		return
	}
	rel, err := h.svc.Release.Publish(r.PathValue("id"), req.ModelID, req.InversionID, req.Permeability, req.IntervalLo, req.IntervalHi, req.Notes)
	if err != nil {
		writeError(w, statusForError(err), err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, rel)
}

func (h *Server) listReleases(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.Release.List(r.PathValue("id"))
	if err != nil {
		writeError(w, statusForError(err), err.Error())
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (h *Server) getRelease(w http.ResponseWriter, r *http.Request) {
	rel, err := h.svc.Release.Get(r.PathValue("id"))
	if err != nil {
		writeError(w, statusForError(err), err.Error())
		return
	}
	writeJSON(w, http.StatusOK, rel)
}
