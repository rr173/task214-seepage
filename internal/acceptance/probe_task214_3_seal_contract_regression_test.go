package acceptance

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"task214-seepage/internal/httpapi"
	"task214-seepage/internal/model"
	"task214-seepage/internal/service"
	"task214-seepage/internal/store"
	"task214-seepage/internal/trial"
)

func TestBug03_SealContractPersistsSealedState(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "seal.db"))
	if err != nil { t.Fatal(err) }
	defer st.Close()
	exp, err := trial.New(st).Create("seal", model.Geometry{Length: 1, Diameter: 0.1, Porosity: 0.35}, model.Fluid{Viscosity: 1e-3, Density: 1000, Compressibility: 1e-9})
	if err != nil { t.Fatal(err) }
	h := service.NewServices(st)
	srv := httptest.NewRecorder()
	httpapi.New(h).Routes().ServeHTTP(srv, httptest.NewRequest(http.MethodPatch, "/api/experiments/"+exp.ID+"/seal", nil))
	if srv.Code != http.StatusOK { t.Fatalf("seal status=%d body=%s", srv.Code, srv.Body.String()) }
	var response map[string]string
	if err := json.Unmarshal(srv.Body.Bytes(), &response); err != nil || response["status"] != "sealed" { t.Fatalf("seal response must say sealed: %s", srv.Body.String()) }
	got, err := trial.New(st).Get(exp.ID)
	if err != nil { t.Fatal(err) }
	if got.State != model.ExpSealed { t.Fatalf("sealed experiment must persist sealed state, got %s", got.State) }
}
