package solver

import (
	"math"
	"testing"
)

func testFluid() FluidGeom {
	return FluidGeom{Length: 1.0, Porosity: 0.35, Viscosity: 1e-3, Compressibility: 1e-9}
}

func testCfg(nt int) SimConfig {
	return SimConfig{NG: 41, NT: nt, T0: 0, T1: 0.3}
}

func ramp() BoundaryCond { return RampBC(1e5, 0.05, 0) }

func obsAt(xfracs []float64, nt int) []Observation {
	o := make([]Observation, 0, len(xfracs)*nt)
	cfg := testCfg(nt)
	for _, xf := range xfracs {
		for n := 1; n <= nt; n++ {
			t := cfg.T0 + float64(n)*0.3/float64(nt)
			o = append(o, Observation{XFrac: xf, T: t})
		}
	}
	return o
}

func TestInvertUniformRecoversUniform(t *testing.T) {
	fg := testFluid()
	bc := ramp()
	cfg := testCfg(300)
	truth := Medium{Layers: []Layer{{Permeability: 2e-12, ThicknessFrac: 1}}}
	obs, err := SimulateObservations(truth, fg, bc, cfg, obsAt([]float64{0.25, 0.5, 0.75}, 60))
	if err != nil {
		t.Fatal(err)
	}
	k, _, _, _ := InvertUniform(fg, bc, cfg, obs, 1e-14, 1e-10)
	if math.Abs(k-2e-12)/2e-12 > 0.1 {
		t.Fatalf("uniform recovery off: got %g want ~2e-12", k)
	}
}

func TestIdentifiabilitySparseVsDense(t *testing.T) {
	fg := testFluid()
	bc := ramp()
	cfg := testCfg(200)
	truth := Medium{Layers: []Layer{
		{Permeability: 1e-12, ThicknessFrac: 0.5},
		{Permeability: 5e-12, ThicknessFrac: 0.5},
	}}

	run := func(xfracs []float64, nt int) Comparison {
		obs, err := SimulateObservations(truth, fg, bc, cfg, obsAt(xfracs, nt))
		if err != nil {
			t.Fatal(err)
		}
		kU, sseU, iU, _ := InvertUniform(fg, bc, cfg, obs, 1e-14, 1e-10)
		k1, k2, sseL, iL, _ := InvertLayered(fg, bc, cfg, obs, 1e-14, 1e-10)
		energy := MeanObservationEnergy(obs)
		count := len(obs)
		results := []ModelResult{
			{Form: "uniform", SSE: sseU, NormalizedSSE: sseU / (float64(count) * energy), PrimaryK: kU, Interval: iU, Converged: true, Fits: FitsWell(sseU, count, energy, 1e-3, 1e-2)},
			{Form: "layered", SSE: sseL, NormalizedSSE: sseL / (float64(count) * energy), PrimaryK: k1, Interval: iL, Converged: true, Fits: FitsWell(sseL, count, energy, 1e-3, 1e-2)},
		}
		t.Logf("xfracs=%v nt=%d | uniform k=%.3g nSSE=%.3g ratio=%.2g | layered k1=%.3g k2=%.3g nSSE=%.3g ratio=%.2g",
			xfracs, nt, kU, sseU/(float64(len(obs))*energy), iU[1]/iU[0],
			k1, k2, sseL/(float64(len(obs))*energy), iL[1]/iL[0])
		return CompareModels(results, 1e-3)
	}

	// Sparse: observing only the outlet (a Dirichlet boundary) lets every
	// candidate reproduce it trivially, so both fit -> non-identifiable.
	sparse := run([]float64{1.0}, 4)
	if sparse.Identifiable {
		t.Errorf("sparse(outlet-only) case should be NON-identifiable: %s", sparse.Reason)
	}
	// Dense: three interior sensors with many samples let only the correct
	// layered model fit well -> identifiable.
	dense := run([]float64{0.25, 0.5, 0.75}, 60)
	if !dense.Identifiable {
		t.Errorf("dense(internal) case should be identifiable: %s", dense.Reason)
	}
}
