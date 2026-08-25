package acceptance

import (
	"path/filepath"
	"testing"

	"task214-seepage/internal/boundary"
	"task214-seepage/internal/model"
	"task214-seepage/internal/store"
	"task214-seepage/internal/trial"
)

func TestBug05_ModelStatePersistsAcrossServiceBoundary(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "model.db"))
	if err != nil { t.Fatal(err) }
	defer st.Close()
	exp, err := trial.New(st).Create("model", model.Geometry{Length: 1, Diameter: 0.1, Porosity: 0.35}, model.Fluid{Viscosity: 1e-3, Density: 1000, Compressibility: 1e-9})
	if err != nil { t.Fatal(err) }
	svc := boundary.New(st)
	m, err := svc.CreateUniform(exp.ID, "uniform", 1e-12)
	if err != nil { t.Fatal(err) }
	if err := svc.SetState(m.ID, model.ModelRunnable); err != nil { t.Fatal(err) }
	got, err := svc.Get(m.ID)
	if err != nil { t.Fatal(err) }
	if got.State != model.ModelRunnable { t.Fatalf("model state must persist as runnable, got %s", got.State) }
}
