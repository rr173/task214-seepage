package inversion

import (
	"testing"

	"task214-seepage/internal/model"
	"task214-seepage/internal/series"
	"task214-seepage/internal/solver"
)

func TestBug07_BoundaryReplayClampsAndSeparatesOutlet(t *testing.T) {
	seq := model.NewSensorSequence("exp", model.SensorPressure, "inlet", 0, "stage", "Pa", []model.Sample{{T: 1, Value: 10}, {T: 2, Value: 20}})
	f := sequenceFunc(seq)
	if f(0) != 10 || f(3) != 20 { t.Fatalf("boundary replay must clamp outside its sample window: got %g and %g", f(0), f(3)) }
	inlet, outlet, obs := series.SplitBoundaryAndObservations([]*model.SensorSequence{seq, model.NewSensorSequence("exp", model.SensorPressure, "outlet", 1, "stage", "Pa", []model.Sample{{T: 1, Value: 0}})})
	if inlet == nil || outlet == nil || len(obs) != 0 { t.Fatalf("inlet/outlet boundaries must not become observations: inlet=%v outlet=%v obs=%d", inlet != nil, outlet != nil, len(obs)) }
	if solver.PiecewiseLinear([]float64{1, 2}, []float64{10, 20})(0) != 10 { t.Fatal("shared interpolation must clamp to the first sample") }
}
