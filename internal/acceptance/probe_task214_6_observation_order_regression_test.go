package acceptance

import (
	"testing"

	"task214-seepage/internal/model"
	"task214-seepage/internal/series"
)

func TestBug06_ObservationOrderIsChronological(t *testing.T) {
	seq := model.NewSensorSequence("exp", model.SensorPressure, "internal", 0.5, "stage", "Pa", []model.Sample{{T: 2, Value: 20}, {T: 0, Value: 0}, {T: 1, Value: 10}})
	obs := series.ToObservations([]*model.SensorSequence{seq})
	if len(obs) != 3 || obs[0].T != 0 || obs[1].T != 1 || obs[2].T != 2 { t.Fatalf("observations must be chronological, got %+v", obs) }
	other := model.NewSensorSequence("exp", model.SensorPressure, "internal", 0.5, "stage", "Pa", []model.Sample{{T: 0, Value: 0}, {T: 1, Value: 10}, {T: 2, Value: 20}})
	if seq.WindowHash != other.WindowHash { t.Fatalf("sample ordering must not change the sequence identity") }
}
