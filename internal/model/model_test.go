package model

import "testing"

func TestExperimentTransitionTable(t *testing.T) {
	e := &Experiment{State: ExpPrepared}
	if !e.CanTransition(ExpSampling) || !e.CanTransition(ExpSealed) {
		t.Fatal("prepared experiment should allow sampling and sealing")
	}
	if e.CanTransition(ExpCompleted) {
		t.Fatal("prepared experiment must not skip directly to completed")
	}
	e.State = ExpSealed
	if e.CanTransition(ExpSampling) {
		t.Fatal("sealed experiment must be terminal")
	}
}

func TestSensorWindowHashIgnoresSampleOrder(t *testing.T) {
	a := NewSensorSequence("exp-1", SensorPressure, "inlet", 0, "stage-1", "Pa", []Sample{{T: 2, Value: 20}, {T: 1, Value: 10}})
	b := NewSensorSequence("exp-1", SensorPressure, "inlet", 0, "stage-1", "Pa", []Sample{{T: 1, Value: 10}, {T: 2, Value: 20}})
	if a.WindowHash != b.WindowHash {
		t.Fatalf("same sample window should hash identically: %s != %s", a.WindowHash, b.WindowHash)
	}
}
