package acceptance

import (
	"path/filepath"
	"testing"

	"task214-seepage/internal/model"
	"task214-seepage/internal/release"
	"task214-seepage/internal/store"
	"task214-seepage/internal/trial"
)

func TestBug01_ReleaseSequenceMonotonic(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "release.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	exp, err := trial.New(st).Create("sequence", model.Geometry{Length: 1, Diameter: 0.1, Porosity: 0.35}, model.Fluid{Viscosity: 1e-3, Density: 1000, Compressibility: 1e-9})
	if err != nil {
		t.Fatal(err)
	}
	svc := release.New(st)
	first, err := svc.Publish(exp.ID, "model-1", "inv-1", 1e-12, 1e-13, 1e-11, "first")
	if err != nil {
		t.Fatal(err)
	}
	second, err := svc.Publish(exp.ID, "model-1", "inv-2", 2e-12, 2e-13, 2e-11, "second")
	if err != nil {
		t.Fatal(err)
	}
	if first.Seq != 1 || second.Seq != 2 {
		t.Fatalf("release sequence must be monotonic from one: got %d then %d", first.Seq, second.Seq)
	}
	items, err := svc.List(exp.ID)
	if err != nil || len(items) != 2 || items[0].Seq != 1 || items[1].Seq != 2 {
		t.Fatalf("persisted release order is wrong: items=%+v err=%v", items, err)
	}
}
