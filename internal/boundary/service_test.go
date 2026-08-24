package boundary

import (
	"path/filepath"
	"testing"

	"task214-seepage/internal/model"
	"task214-seepage/internal/store"
	"task214-seepage/internal/trial"
)

func TestSeedDefaultsCreatesRunnableCandidates(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "boundary.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	exp, err := trial.New(st).Create("column", model.Geometry{Length: 1, Diameter: 0.1, Porosity: 0.35}, model.Fluid{Viscosity: 1e-3, Density: 1000, Compressibility: 1e-9})
	if err != nil {
		t.Fatal(err)
	}
	pair, err := New(st).SeedDefaults(exp.ID, 1e-12, 5e-12)
	if err != nil {
		t.Fatal(err)
	}
	if pair.Uniform.Form != model.ModelUniform || pair.Layered.Form != model.ModelLayered {
		t.Fatalf("unexpected candidate forms: %s and %s", pair.Uniform.Form, pair.Layered.Form)
	}
	for _, candidate := range []*model.BoundaryModel{pair.Uniform, pair.Layered} {
		if candidate.State != model.ModelRunnable {
			t.Fatalf("candidate %s should be runnable, got %s", candidate.ID, candidate.State)
		}
	}
}
