package acceptance

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"task214-seepage/internal/httpapi"
	"task214-seepage/internal/service"
	"task214-seepage/internal/store"
)

func TestBug09_MissingExperimentMapsToNotFound(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "notfound.db"))
	if err != nil { t.Fatal(err) }
	defer st.Close()
	rec := httptest.NewRecorder()
	httpapi.New(service.NewServices(st)).Routes().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/experiments/missing", nil))
	if rec.Code != http.StatusNotFound { t.Fatalf("missing experiment must return 404, got %d body=%s", rec.Code, rec.Body.String()) }
}
