package series

import (
	"errors"
	"path/filepath"
	"testing"

	"task214-seepage/internal/model"
	"task214-seepage/internal/store"
	"task214-seepage/internal/trial"
)

func TestUploadIdempotencyAndAssessment(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "series.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	exp, err := trial.New(st).Create("column", model.Geometry{Length: 1, Diameter: 0.1, Porosity: 0.35}, model.Fluid{Viscosity: 1e-3, Density: 1000, Compressibility: 1e-9})
	if err != nil {
		t.Fatal(err)
	}
	svc := New(st)
	samples := []model.Sample{{T: 0, Value: 1}, {T: 1, Value: 2}, {T: 2, Value: 3}}
	first, err := svc.Upload(exp.ID, "pressure", "inlet", 0, "stage-1", "Pa", samples)
	if err != nil {
		t.Fatal(err)
	}
	if first.State != model.SeqValid {
		t.Fatalf("expected valid sequence, got %s", first.State)
	}
	second, err := svc.Upload(exp.ID, "pressure", "inlet", 0, "stage-1", "Pa", samples)
	if !errors.Is(err, model.ErrDuplicateSequence) {
		t.Fatalf("duplicate upload should return ErrDuplicateSequence, got %v", err)
	}
	if second == nil || second.ID != first.ID {
		gotID := "<nil>"
		if second != nil {
			gotID = second.ID
		}
		t.Fatalf("duplicate upload should return existing record, id=%q", gotID)
	}
	if err := Validate(model.NewSensorSequence(exp.ID, model.SensorFlow, "flow", 0.5, "stage-1", "m3/s", []model.Sample{{T: 0, Value: -1}})); !errors.Is(err, model.ErrNegativeFlow) {
		t.Fatalf("negative flow should be rejected, got %v", err)
	}
}
