package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"task214-seepage/internal/service"
	"task214-seepage/internal/store"
)

func TestRoutesCreateExperimentAndSelfCheck(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "http.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	h := New(service.NewServices(st)).Routes()
	body := bytes.NewBufferString(`{"name":"http-column","length_m":1,"diameter_m":0.1,"porosity":0.35,"viscosity_pas":0.001,"density_kgm3":1000,"compressibility_pa":1e-9}`)
	req := httptest.NewRequest(http.MethodPost, "/api/experiments", body)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create experiment status=%d body=%s", rec.Code, rec.Body.String())
	}
	var created struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil || created.ID == "" {
		t.Fatalf("invalid create response: %s err=%v", rec.Body.String(), err)
	}
	check := httptest.NewRequest(http.MethodGet, "/api/selfcheck", nil)
	checkRec := httptest.NewRecorder()
	h.ServeHTTP(checkRec, check)
	if checkRec.Code != http.StatusOK {
		t.Fatalf("selfcheck status=%d body=%s", checkRec.Code, checkRec.Body.String())
	}
}
