package series

import (
	"errors"
	"path/filepath"
	"testing"

	"task214-seepage/internal/model"
	"task214-seepage/internal/store"
	"task214-seepage/internal/trial"
)

// TestCrossExperimentWindowIndependence guards the sequence-identity fix: the
// same window of data filed against two different experiments must produce two
// independent sequence records (distinct ids and distinct window hashes), and
// must not be mistaken for a duplicate of one another. Idempotency still holds
// within a single experiment for an exact re-upload.
func TestCrossExperimentWindowIndependence(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "cross.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	tr := trial.New(st)
	expA, err := tr.Create("column-A", model.Geometry{Length: 1, Diameter: 0.1, Porosity: 0.35}, model.Fluid{Viscosity: 1e-3, Density: 1000, Compressibility: 1e-9})
	if err != nil {
		t.Fatal(err)
	}
	expB, err := tr.Create("column-B", model.Geometry{Length: 1, Diameter: 0.1, Porosity: 0.35}, model.Fluid{Viscosity: 1e-3, Density: 1000, Compressibility: 1e-9})
	if err != nil {
		t.Fatal(err)
	}
	svc := New(st)

	// Identical window data filed against two different experiments.
	samples := []model.Sample{{T: 0, Value: 1}, {T: 1, Value: 2}, {T: 2, Value: 3}}
	a, err := svc.Upload(expA.ID, "pressure", "inlet", 0, "stage-1", "Pa", samples)
	if err != nil {
		t.Fatalf("upload A: %v", err)
	}
	b, err := svc.Upload(expB.ID, "pressure", "inlet", 0, "stage-1", "Pa", samples)
	if err != nil {
		t.Fatalf("upload B: %v", err)
	}

	if a.ID == b.ID {
		t.Fatalf("cross-experiment uploads collapsed onto the same record id %q", a.ID)
	}
	if a.WindowHash == b.WindowHash {
		t.Fatalf("cross-experiment uploads share a window hash %q; experiment id must participate in the hash", a.WindowHash)
	}

	// Each experiment must see exactly its own sequence via List.
	listA, err := svc.List(expA.ID)
	if err != nil {
		t.Fatal(err)
	}
	listB, err := svc.List(expB.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(listA) != 1 || listA[0].ID != a.ID {
		t.Fatalf("experiment A should own only its sequence, got %+v", listA)
	}
	if len(listB) != 1 || listB[0].ID != b.ID {
		t.Fatalf("experiment B should own only its sequence, got %+v", listB)
	}

	// Exact re-upload within the same experiment stays idempotent.
	dup, err := svc.Upload(expA.ID, "pressure", "inlet", 0, "stage-1", "Pa", samples)
	if !errors.Is(err, model.ErrDuplicateSequence) || dup == nil || dup.ID != a.ID {
		t.Fatalf("same-experiment re-upload should return the existing record with ErrDuplicateSequence, got dup=%v err=%v", dup, err)
	}
}
