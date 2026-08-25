package acceptance

import (
	"path/filepath"
	"testing"

	"task214-seepage/internal/model"
	"task214-seepage/internal/release"
	"task214-seepage/internal/store"
	"task214-seepage/internal/trial"
)

func TestBug02_ReleaseCompletesPendingExperiment(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "completion.db"))
	if err != nil { t.Fatal(err) }
	defer st.Close()
	exp, err := trial.New(st).Create("completion", model.Geometry{Length: 1, Diameter: 0.1, Porosity: 0.35}, model.Fluid{Viscosity: 1e-3, Density: 1000, Compressibility: 1e-9})
	if err != nil { t.Fatal(err) }
	if err := trial.New(st).Transition(exp.ID, model.ExpSampling); err != nil { t.Fatal(err) }
	if err := trial.New(st).Transition(exp.ID, model.ExpPendingInversion); err != nil { t.Fatal(err) }
	if _, err := release.New(st).Publish(exp.ID, "model-1", "inv-1", 1e-12, 1e-13, 1e-11, "published"); err != nil { t.Fatal(err) }
	got, err := trial.New(st).Get(exp.ID)
	if err != nil { t.Fatal(err) }
	if got.State != model.ExpCompleted { t.Fatalf("publishing a pending inversion should complete the experiment, got %s", got.State) }
}
