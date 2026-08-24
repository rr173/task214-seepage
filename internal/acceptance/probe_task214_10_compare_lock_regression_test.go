package acceptance

import (
	"path/filepath"
	"sync"
	"testing"

	"task214-seepage/internal/boundary"
	"task214-seepage/internal/model"
	"task214-seepage/internal/service"
	"task214-seepage/internal/store"
	"task214-seepage/internal/trial"
)

func TestBug10_ConcurrentCompareUsesSharedModelLocks(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "compare.db"))
	if err != nil { t.Fatal(err) }
	defer st.Close()
	exp, err := trial.New(st).Create("compare", model.Geometry{Length: 1, Diameter: 0.1, Porosity: 0.35}, model.Fluid{Viscosity: 1e-3, Density: 1000, Compressibility: 1e-9})
	if err != nil { t.Fatal(err) }
	svc := service.NewServices(st)
	if _, err := svc.Series.Upload(exp.ID, "pressure", "inlet", 0, "stage-1", "Pa", []model.Sample{{T: 0, Value: 1}, {T: 1, Value: 2}}); err != nil { t.Fatal(err) }
	if _, err := svc.Series.Upload(exp.ID, "pressure", "internal", 0.5, "stage-1", "Pa", []model.Sample{{T: 0.1, Value: 0.5}, {T: 0.2, Value: 0.6}}); err != nil { t.Fatal(err) }
	m, err := boundary.New(st).CreateUniform(exp.ID, "uniform", 1e-12)
	if err != nil { t.Fatal(err) }
	if err := boundary.New(st).SetState(m.ID, model.ModelRunnable); err != nil { t.Fatal(err) }
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := svc.Compare.Compare(exp.ID); err != nil { t.Errorf("compare failed: %v", err) }
		}()
	}
	wg.Wait()
}
