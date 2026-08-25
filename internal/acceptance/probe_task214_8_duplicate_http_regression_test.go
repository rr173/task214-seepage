package acceptance

import (
	"bytes"
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

func TestBug08_DuplicateSequenceHTTPReturnsExistingRecord(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "duplicate.db"))
	if err != nil { t.Fatal(err) }
	defer st.Close()
	exp, err := trial.New(st).Create("duplicate", model.Geometry{Length: 1, Diameter: 0.1, Porosity: 0.35}, model.Fluid{Viscosity: 1e-3, Density: 1000, Compressibility: 1e-9})
	if err != nil { t.Fatal(err) }
	body := []byte(`{"sensor_type":"pressure","location":"inlet","x_frac":0,"stage":"stage-1","unit":"Pa","samples":[{"t":0,"value":1},{"t":1,"value":2}]}`)
	routes := httpapi.New(service.NewServices(st)).Routes()
	for i := 0; i < 2; i++ {
		rec := httptest.NewRecorder()
		routes.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/experiments/"+exp.ID+"/sequences", bytes.NewReader(body)))
		want := http.StatusCreated
		if i == 1 { want = http.StatusOK }
		if rec.Code != want { t.Fatalf("upload %d status=%d body=%s", i+1, rec.Code, rec.Body.String()) }
	}
}
