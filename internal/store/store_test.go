package store

import (
	"errors"
	"path/filepath"
	"testing"

	"task214-seepage/internal/model"
)

func TestExperimentAndSequenceSurviveReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "store.db")
	st, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	exp := model.NewExperiment("persisted", model.Geometry{Length: 1, Diameter: 0.1, Porosity: 0.35}, model.Fluid{Viscosity: 1e-3, Density: 1000, Compressibility: 1e-9})
	if err := st.CreateExperiment(exp); err != nil {
		t.Fatal(err)
	}
	seq := model.NewSensorSequence(exp.ID, model.SensorPressure, "inlet", 0, "stage-1", "Pa", []model.Sample{{T: 0, Value: 1}})
	if _, err := st.CreateSensorSequence(seq); err != nil {
		t.Fatal(err)
	}
	if err := st.Close(); err != nil {
		t.Fatal(err)
	}
	st, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	got, err := st.GetExperiment(exp.ID)
	if err != nil || got.Name != exp.Name {
		t.Fatalf("experiment did not survive reopen: got=%+v err=%v", got, err)
	}
	seqs, err := st.ListSensorSequences(exp.ID)
	if err != nil || len(seqs) != 1 || seqs[0].WindowHash != seq.WindowHash {
		t.Fatalf("sequence did not survive reopen: got=%+v err=%v", seqs, err)
	}
	if _, err := st.GetExperiment("missing"); !errors.Is(err, model.ErrNotFound) {
		t.Fatalf("missing experiment should return ErrNotFound, got %v", err)
	}
}
