package acceptance

import (
	"path/filepath"
	"testing"

	"task214-seepage/internal/model"
	"task214-seepage/internal/series"
	"task214-seepage/internal/store"
	"task214-seepage/internal/trial"
)

func TestBug04_SequenceIdentityIncludesExperiment(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "scope.db"))
	if err != nil { t.Fatal(err) }
	defer st.Close()
	makeExp := func(name string) string {
		e, err := trial.New(st).Create(name, model.Geometry{Length: 1, Diameter: 0.1, Porosity: 0.35}, model.Fluid{Viscosity: 1e-3, Density: 1000, Compressibility: 1e-9})
		if err != nil { t.Fatal(err) }
		return e.ID
	}
	a, b := makeExp("a"), makeExp("b")
	samples := []model.Sample{{T: 0, Value: 1}, {T: 1, Value: 2}}
	first, err := series.New(st).Upload(a, "pressure", "inlet", 0, "stage-1", "Pa", samples)
	if err != nil { t.Fatal(err) }
	second, err := series.New(st).Upload(b, "pressure", "inlet", 0, "stage-1", "Pa", samples)
	if err != nil { t.Fatal(err) }
	if first.ID == second.ID { t.Fatalf("same sensor window in different experiments must have different records: %s", first.ID) }
}
