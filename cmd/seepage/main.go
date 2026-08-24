// Command seepage is the entry point for the porous-media seepage experiment
// boundary-inversion service. In normal mode it serves the HTTP API; with
// --smoke-test it exercises the full business loop against a throwaway database
// and verifies persistence across a close/reopen cycle.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"task214-seepage/internal/httpapi"
	"task214-seepage/internal/model"
	"task214-seepage/internal/service"
	"task214-seepage/internal/solver"
	"task214-seepage/internal/store"
)

func main() {
	addr := flag.String("addr", ":8080", "HTTP listen address")
	dbPath := flag.String("db", "seepage.db", "SQLite database path")
	smoke := flag.Bool("smoke-test", false, "run the smoke test and exit")
	flag.Parse()

	if *smoke {
		if err := runSmokeTest(*dbPath); err != nil {
			fmt.Fprintln(os.Stderr, "SMOKE-TEST FAILED:", err)
			os.Exit(1)
		}
		fmt.Println("SMOKE-TEST PASSED")
		os.Exit(0)
	}
	if err := runServer(*addr, *dbPath); err != nil {
		log.Fatal(err)
	}
}

func runServer(addr, dbPath string) error {
	st, err := store.Open(dbPath)
	if err != nil {
		return err
	}
	defer st.Close()
	svc := service.NewServices(st)
	srv := &http.Server{Addr: addr, Handler: httpapi.New(svc).Routes()}
	log.Printf("seepage boundary-inversion service listening on %s", addr)
	return srv.ListenAndServe()
}

func linspace(lo, hi float64, n int) []float64 {
	out := make([]float64, n)
	for i := 0; i < n; i++ {
		out[i] = lo + (hi-lo)*float64(i)/float64(n-1)
	}
	return out
}

// runSmokeTest drives the end-to-end loop and verifies restart recovery.
func runSmokeTest(_ string) error {
	tmp, err := os.MkdirTemp("", "seepage-smoke-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	path := filepath.Join(tmp, "smoke.db")

	st, err := store.Open(path)
	if err != nil {
		return err
	}
	svc := service.NewServices(st)

	// 1. experiment + lifecycle
	exp, err := svc.Trial.Create("smoke-column",
		model.Geometry{Length: 1.0, Diameter: 0.1, Porosity: 0.35},
		model.Fluid{Viscosity: 1e-3, Density: 1000, Compressibility: 1e-9})
	if err != nil {
		return err
	}
	for _, to := range []model.ExperimentState{model.ExpSampling, model.ExpPendingInversion} {
		if err := svc.Trial.Transition(exp.ID, to); err != nil {
			return err
		}
	}

	// 2. numerical setup and true (layered) medium
	fg := solver.FluidGeom{Length: 1.0, Porosity: 0.35, Viscosity: 1e-3, Compressibility: 1e-9}
	cfg := solver.SimConfig{NG: 41, NT: 200, T0: 0, T1: 0.3}
	truth := solver.Medium{Layers: []solver.Layer{
		{Permeability: 1e-12, ThicknessFrac: 0.5},
		{Permeability: 5e-12, ThicknessFrac: 0.5},
	}}
	times := linspace(0.02, 0.3, 60)

	// inlet boundary: sample the physical ramp and replay it through the same
	// piecewise-linear reconstruction the inversion uses, so the data generator
	// and the boundary reconstruction agree exactly (SSE is 0 at the truth).
	ramp := solver.RampBC(1e5, 0.05, 0)
	inletVals := make([]float64, len(times))
	for i, t := range times {
		inletVals[i] = ramp.InletPressure(t)
	}
	bc := solver.BoundaryCond{
		InletPressure:  solver.PiecewiseLinear(times, inletVals),
		OutletPressure: func(t float64) float64 { return 0 },
	}
	field, err := solver.Forward(truth, fg, bc, cfg)
	if err != nil {
		return err
	}

	// 3. inlet pressure boundary condition (same samples as the reconstruction)
	inlet := make([]model.Sample, len(times))
	for i, t := range times {
		inlet[i] = model.Sample{T: t, Value: inletVals[i]}
	}
	if _, err := svc.Series.Upload(exp.ID, "pressure", "inlet", 0, "stage-1", "Pa", inlet); err != nil {
		return err
	}

	// 4. SPARSE observations: only an outlet survey (a Dirichlet boundary that
	// every candidate reproduces trivially) -> non-identifiable.
	outlet := make([]model.Sample, len(times))
	for i, t := range times {
		outlet[i] = model.Sample{T: t, Value: solver.SampleField(field, cfg, 1.0, t)}
	}
	if _, err := svc.Series.Upload(exp.ID, "pressure", "outlet-survey", 1.0, "stage-1", "Pa", outlet); err != nil {
		return err
	}

	// 5. candidate models
	pair, err := svc.Boundary.SeedDefaults(exp.ID, 1e-12, 5e-12)
	if err != nil {
		return err
	}

	// 6. compare on sparse data -> expect non-identifiable
	res, err := svc.Compare.Compare(exp.ID)
	if err != nil {
		return err
	}
	if res.Comparison.Identifiable {
		return fmt.Errorf("sparse: expected non-identifiable, got identifiable: %s", res.Comparison.Reason)
	}
	log.Printf("sparse comparison: non-identifiable (%d tasks)", len(res.Tasks))

	// 7. DENSE observations: add internal sensors -> identifiable
	for _, xf := range []float64{0.25, 0.5, 0.75} {
		samples := make([]model.Sample, len(times))
		for i, t := range times {
			samples[i] = model.Sample{T: t, Value: solver.SampleField(field, cfg, xf, t)}
		}
		if _, err := svc.Series.Upload(exp.ID, "pressure", fmt.Sprintf("internal-%.2f", xf), xf, "stage-1", "Pa", samples); err != nil {
			return err
		}
	}
	res2, err := svc.Compare.Compare(exp.ID)
	if err != nil {
		return err
	}
	if !res2.Comparison.Identifiable {
		return fmt.Errorf("dense: expected identifiable, got non-identifiable: %s", res2.Comparison.Reason)
	}
	log.Printf("dense comparison: identifiable (%s)", res2.Comparison.Reason)

	// 8. run inversion on the correct (layered) model and publish a release
	task, err := svc.Inversion.Run(pair.Layered.ID)
	if err != nil {
		return err
	}
	if task.State != model.InvConverged {
		return fmt.Errorf("layered inversion did not converge: %s", task.State)
	}
	if _, err := svc.Release.Publish(exp.ID, pair.Layered.ID, task.ID, task.EstimatedLayers[0].Permeability, task.IdentifiabilityLo, task.IdentifiabilityHi, "smoke release"); err != nil {
		return err
	}

	// 9. CLOSE and REOPEN the same database to verify restart recovery
	st.Close()
	time.Sleep(50 * time.Millisecond)
	st2, err := store.Open(path)
	if err != nil {
		return err
	}
	defer st2.Close()
	svc2 := service.NewServices(st2)

	exp2, err := svc2.Trial.Get(exp.ID)
	if err != nil || exp2 == nil {
		return fmt.Errorf("restart recovery failed: experiment not found")
	}
	seqs, _ := svc2.Series.List(exp.ID)
	if len(seqs) != 5 {
		return fmt.Errorf("restart recovery: expected 5 sequences, got %d", len(seqs))
	}
	models, _ := svc2.Boundary.List(exp.ID)
	if len(models) != 2 {
		return fmt.Errorf("restart recovery: expected 2 models, got %d", len(models))
	}
	invs, _ := svc2.Inversion.ListTasks(exp.ID)
	if len(invs) == 0 {
		return fmt.Errorf("restart recovery: no inversion tasks persisted")
	}
	rels, _ := svc2.Release.List(exp.ID)
	if len(rels) != 1 {
		return fmt.Errorf("restart recovery: expected 1 release, got %d", len(rels))
	}
	log.Printf("restart recovery verified: exp=%s seqs=%d models=%d inversions=%d releases=%d",
		exp2.ID, len(seqs), len(models), len(invs), len(rels))
	return nil
}
