package trial

import (
	"errors"
	"path/filepath"
	"testing"

	"task214-seepage/internal/model"
	"task214-seepage/internal/store"
)

func TestServiceLifecycleAndSeal(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "trial.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	svc := New(st)
	e, err := svc.Create("column", model.Geometry{Length: 1, Diameter: 0.1, Porosity: 0.35}, model.Fluid{Viscosity: 1e-3, Density: 1000, Compressibility: 1e-9})
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.Transition(e.ID, model.ExpPendingInversion); !errors.Is(err, model.ErrInvalidState) {
		t.Fatalf("skipping lifecycle state should fail with ErrInvalidState, got %v", err)
	}
	if err := svc.Transition(e.ID, model.ExpSampling); err != nil {
		t.Fatal(err)
	}
	if err := svc.Seal(e.ID); err != nil {
		t.Fatal(err)
	}
	if err := svc.Transition(e.ID, model.ExpPendingInversion); !errors.Is(err, model.ErrSealed) {
		t.Fatalf("sealed experiment should reject transitions, got %v", err)
	}
}
