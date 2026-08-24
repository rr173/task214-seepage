package httpapi

import (
	"errors"
	"net/http"

	"task214-seepage/internal/model"
)

type uploadSequenceReq struct {
	SensorType string          `json:"sensor_type"`
	Location   string          `json:"location"`
	XFrac      float64         `json:"x_frac"`
	Stage      string          `json:"stage"`
	Unit       string          `json:"unit"`
	Samples    []model.Sample  `json:"samples"`
}

func (h *Server) uploadSequence(w http.ResponseWriter, r *http.Request) {
	var req uploadSequenceReq
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body: "+err.Error())
		return
	}
	seq, err := h.svc.Series.Upload(r.PathValue("id"), req.SensorType, req.Location, req.XFrac, req.Stage, req.Unit, req.Samples)
	if err != nil {
		if errors.Is(err, model.ErrDuplicateSequence) {
			writeJSON(w, http.StatusOK, map[string]interface{}{"id": seq.ID, "state": string(seq.State), "idempotent": true})
			return
		}
		writeError(w, statusForError(err), err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, seq)
}

func (h *Server) listSequences(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.Series.List(r.PathValue("id"))
	if err != nil {
		writeError(w, statusForError(err), err.Error())
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (h *Server) getSequence(w http.ResponseWriter, r *http.Request) {
	seq, err := h.svc.Series.Get(r.PathValue("id"))
	if err != nil {
		writeError(w, statusForError(err), err.Error())
		return
	}
	writeJSON(w, http.StatusOK, seq)
}

func (h *Server) revalidateSequence(w http.ResponseWriter, r *http.Request) {
	seq, err := h.svc.Series.Revalidate(r.PathValue("id"))
	if err != nil {
		writeError(w, statusForError(err), err.Error())
		return
	}
	writeJSON(w, http.StatusOK, seq)
}

func (h *Server) sequenceStats(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.Series.List(r.PathValue("id"))
	if err != nil {
		writeError(w, statusForError(err), err.Error())
		return
	}
	stats := map[string]interface{}{
		"total":          len(list),
		"by_type":        map[string]int{},
		"by_state":       map[string]int{},
		"total_samples":  0,
	}
	byType := stats["by_type"].(map[string]int)
	byState := stats["by_state"].(map[string]int)
	samples := 0
	for _, s := range list {
		byType[string(s.SensorType)]++
		byState[string(s.State)]++
		samples += len(s.Samples)
	}
	stats["total_samples"] = samples
	writeJSON(w, http.StatusOK, stats)
}

type createModelReq struct {
	Form  string  `json:"form"`
	Name  string  `json:"name"`
	K     float64 `json:"k"`
	K1    float64 `json:"k1"`
	K2    float64 `json:"k2"`
}

func (h *Server) createModel(w http.ResponseWriter, r *http.Request) {
	var req createModelReq
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body: "+err.Error())
		return
	}
	var (
		m   *model.BoundaryModel
		err error
	)
	switch model.ModelForm(req.Form) {
	case model.ModelUniform:
		m, err = h.svc.Boundary.CreateUniform(r.PathValue("id"), req.Name, req.K)
	case model.ModelLayered:
		m, err = h.svc.Boundary.CreateLayered(r.PathValue("id"), req.Name, req.K1, req.K2)
	default:
		writeError(w, http.StatusBadRequest, "form must be 'uniform' or 'layered'")
		return
	}
	if err != nil {
		writeError(w, statusForError(err), err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, m)
}

func (h *Server) listModels(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.Boundary.List(r.PathValue("id"))
	if err != nil {
		writeError(w, statusForError(err), err.Error())
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (h *Server) getModel(w http.ResponseWriter, r *http.Request) {
	m, err := h.svc.Boundary.Get(r.PathValue("id"))
	if err != nil {
		writeError(w, statusForError(err), err.Error())
		return
	}
	writeJSON(w, http.StatusOK, m)
}

type setModelStatusReq struct {
	Status string `json:"status"`
}

func (h *Server) setModelStatus(w http.ResponseWriter, r *http.Request) {
	var req setModelStatusReq
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body: "+err.Error())
		return
	}
	if err := h.svc.Boundary.SetState(r.PathValue("id"), model.ModelState(req.Status)); err != nil {
		writeError(w, statusForError(err), err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": req.Status, "id": r.PathValue("id")})
}
