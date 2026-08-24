// Package solver implements a one-dimensional transient Darcy pressure-diffusion
// forward model for a porous column together with boundary-parameter inversion
// and candidate-model identifiability analysis.
//
// Governing equation (pressure diffusion):
//
//	∂p/∂t = D(x) · ∂²p/∂x² ,  D(x) = k(x) / (φ · μ · c_t)
//
// where k is permeability, φ porosity, μ dynamic viscosity and c_t total
// compressibility. The column is discretised on NG nodes; an implicit
// (backward-Euler) scheme is solved each time step with a Thomas algorithm, so
// the integration is unconditionally stable for any time step.
package solver

import (
	"errors"
	"math"
	"sort"
)

// Layer is one homogeneous sub-region of a (possibly layered) medium.
type Layer struct {
	Permeability  float64
	ThicknessFrac float64
}

// Medium is the spatial distribution of permeability along the column.
type Medium struct {
	Layers []Layer
}

// PermeabilityAt returns the permeability at fractional position xf in [0,1].
func (m Medium) PermeabilityAt(xf float64) float64 {
	if len(m.Layers) == 0 {
		return 0
	}
	if len(m.Layers) == 1 {
		return m.Layers[0].Permeability
	}
	cum := 0.0
	for _, l := range m.Layers {
		cum += l.ThicknessFrac
		if xf <= cum {
			return l.Permeability
		}
	}
	return m.Layers[len(m.Layers)-1].Permeability
}

// FluidGeom bundles geometry and fluid properties of the experiment.
type FluidGeom struct {
	Length          float64
	Porosity        float64
	Viscosity       float64
	Compressibility float64
}

// Diffusivity returns D = k / (φ·μ·c_t) for a given permeability.
func (fg FluidGeom) Diffusivity(k float64) float64 {
	denom := fg.Porosity * fg.Viscosity * fg.Compressibility
	if denom <= 0 {
		return 0
	}
	return k / denom
}

// BoundaryCond supplies the Dirichlet pressures at the two column ends.
type BoundaryCond struct {
	InletPressure  func(t float64) float64
	OutletPressure func(t float64) float64
}

// SimConfig controls the space/time discretisation.
type SimConfig struct {
	NG int     // number of grid nodes (>=2)
	NT int     // number of time steps (>=1)
	T0 float64 // start time (s)
	T1 float64 // end time (s)
}

// RampBC builds a smooth ramp inlet pressure with a fixed outlet pressure.
func RampBC(pMax, tau, pOut float64) BoundaryCond {
	return BoundaryCond{
		InletPressure: func(t float64) float64 {
			if t <= 0 {
				return 0
			}
			return pMax * (1 - math.Exp(-t/tau))
		},
		OutletPressure: func(t float64) float64 { return pOut },
	}
}

// ConstantBC builds a steady pressure differential between the two ends.
func ConstantBC(pIn, pOut float64) BoundaryCond {
	return BoundaryCond{
		InletPressure:  func(t float64) float64 { return pIn },
		OutletPressure: func(t float64) float64 { return pOut },
	}
}

// Forward integrates the pressure field and returns field[i][n] = pressure at
// node i and time step n. Node 0 is the inlet, node NG-1 the outlet.
func Forward(m Medium, fg FluidGeom, bc BoundaryCond, cfg SimConfig) ([][]float64, error) {
	if cfg.NG < 2 || cfg.NT < 1 {
		return nil, errors.New("solver: NG must be >=2 and NT >=1")
	}
	ng, nt := cfg.NG, cfg.NT
	if fg.Length <= 0 {
		return nil, errors.New("solver: non-positive column length")
	}
	dx := fg.Length / float64(ng-1)
	dt := (cfg.T1 - cfg.T0) / float64(nt)
	if dx <= 0 || dt <= 0 {
		return nil, errors.New("solver: non-positive dx or dt")
	}
	D := make([]float64, ng)
	for i := 0; i < ng; i++ {
		xf := float64(i) / float64(ng-1)
		D[i] = fg.Diffusivity(m.PermeabilityAt(xf))
	}
	field := make([][]float64, ng)
	for i := range field {
		field[i] = make([]float64, nt+1)
	}
	p0 := bc.InletPressure(cfg.T0)
	for i := 0; i < ng; i++ {
		field[i][0] = p0
	}
	for n := 1; n <= nt; n++ {
		t := cfg.T0 + float64(n)*dt
		pin := bc.InletPressure(t)
		pout := bc.OutletPressure(t)
		a := make([]float64, ng)
		b := make([]float64, ng)
		c := make([]float64, ng)
		d := make([]float64, ng)
		for i := 0; i < ng; i++ {
			if i == 0 {
				b[i], c[i], d[i] = 1, 0, pin
				continue
			}
			if i == ng-1 {
				a[i], b[i], d[i] = 0, 1, pout
				continue
			}
			Dm := 0.5 * (D[i-1] + D[i]) // D at i-1/2
			Dp := 0.5 * (D[i] + D[i+1]) // D at i+1/2
			r := dt / (dx * dx)
			a[i] = -r * Dm
			c[i] = -r * Dp
			b[i] = 1 + r*(Dm+Dp)
			d[i] = field[i][n-1]
		}
		sol, err := thomas(a, b, c, d)
		if err != nil {
			return nil, err
		}
		for i := 0; i < ng; i++ {
			field[i][n] = sol[i]
		}
	}
	return field, nil
}

// thomas solves a tridiagonal system a_i x_{i-1} + b_i x_i + c_i x_{i+1} = d_i.
func thomas(a, b, c, d []float64) ([]float64, error) {
	n := len(b)
	cp := make([]float64, n)
	dp := make([]float64, n)
	cp[0] = c[0] / b[0]
	dp[0] = d[0] / b[0]
	for i := 1; i < n; i++ {
		denom := b[i] - a[i]*cp[i-1]
		if denom == 0 {
			return nil, errors.New("solver: singular tridiagonal matrix")
		}
		cp[i] = c[i] / denom
		dp[i] = (d[i] - a[i]*dp[i-1]) / denom
	}
	x := make([]float64, n)
	x[n-1] = dp[n-1]
	for i := n - 2; i >= 0; i-- {
		x[i] = dp[i] - cp[i]*x[i+1]
	}
	return x, nil
}

// SampleField bilinearly interpolates the simulated field at fractional column
// position xf in [0,1] and time t.
func SampleField(field [][]float64, cfg SimConfig, xf, t float64) float64 {
	ng := len(field)
	nt := len(field[0]) - 1
	if ng == 0 || nt < 0 {
		return 0
	}
	xi := xf * float64(ng-1)
	i0 := int(math.Floor(xi))
	if i0 >= ng-1 {
		i0 = ng - 2
	}
	if i0 < 0 {
		i0 = 0
	}
	frac := xi - float64(i0)

	span := cfg.T1 - cfg.T0
	if span <= 0 {
		span = 1
	}
	tn := (t - cfg.T0) / span * float64(nt)
	n0 := int(math.Floor(tn))
	if n0 >= nt {
		n0 = nt - 1
	}
	if n0 < 0 {
		n0 = 0
	}
	tf := tn - float64(n0)

	p00 := field[i0][n0]
	p01 := field[i0][n0+1]
	p10 := field[i0+1][n0]
	p11 := field[i0+1][n0+1]
	pi0 := p00*(1-tf) + p01*tf
	pi1 := p10*(1-tf) + p11*tf
	return pi0*(1-frac) + pi1*frac
}

// Observation is a single measurement used to fit the model.
type Observation struct {
	XFrac float64
	T     float64
	Value float64
}

// SimulateObservations evaluates the forward model at the requested sensor
// locations/times and returns them as observations (used to synthesize data).
func SimulateObservations(m Medium, fg FluidGeom, bc BoundaryCond, cfg SimConfig, sensors []Observation) ([]Observation, error) {
	field, err := Forward(m, fg, bc, cfg)
	if err != nil {
		return nil, err
	}
	out := make([]Observation, len(sensors))
	for i, s := range sensors {
		out[i] = Observation{XFrac: s.XFrac, T: s.T, Value: SampleField(field, cfg, s.XFrac, s.T)}
	}
	return out, nil
}

// cost evaluates the sum of squared residuals between predictions and obs.
func cost(m Medium, fg FluidGeom, bc BoundaryCond, cfg SimConfig, obs []Observation) float64 {
	field, err := Forward(m, fg, bc, cfg)
	if err != nil {
		return math.Inf(1)
	}
	sse := 0.0
	for _, o := range obs {
		pred := SampleField(field, cfg, o.XFrac, o.T)
		d := pred - o.Value
		sse += d * d
	}
	return sse
}

// goldenSection minimises f on [a,b] using the golden-ratio bisection.
func goldenSection(f func(float64) float64, a, b float64, iters int) (float64, float64) {
	gr := (math.Sqrt(5) - 1) / 2
	c := b - gr*(b-a)
	d := a + gr*(b-a)
	fc, fd := f(c), f(d)
	for i := 0; i < iters; i++ {
		if fc < fd {
			b, d, fd = d, c, fc
			c = b - gr*(b-a)
			fc = f(c)
		} else {
			a, c, fc = c, d, fd
			d = a + gr*(b-a)
			fd = f(d)
		}
	}
	if fc < fd {
		return c, fc
	}
	return d, fd
}

// identifiabilityInterval scans the parameter range and returns the [lo,hi]
// permeability interval whose cost stays within tolerance of the best cost.
func identifiabilityInterval(f func(float64) float64, lkBest, sseBest, lo, hi float64) [2]float64 {
	tol := sseBest*0.05 + 1e-9
	thresh := sseBest + tol
	best, worst := lkBest, lkBest
	steps := 400
	for i := 0; i <= steps; i++ {
		x := lo + (hi-lo)*float64(i)/float64(steps)
		if f(x) <= thresh {
			if x < best {
				best = x
			}
			if x > worst {
				worst = x
			}
		}
	}
	return [2]float64{math.Pow(10, best), math.Pow(10, worst)}
}

// InvertUniform estimates a single homogeneous permeability.
func InvertUniform(fg FluidGeom, bc BoundaryCond, cfg SimConfig, obs []Observation, kLo, kHi float64) (k, sse float64, interval [2]float64, converged bool) {
	f := func(lk float64) float64 {
		m := Medium{Layers: []Layer{{Permeability: math.Pow(10, lk), ThicknessFrac: 1}}}
		return cost(m, fg, bc, cfg, obs)
	}
	lk, best := goldenSection(f, math.Log10(kLo), math.Log10(kHi), 200)
	interval = identifiabilityInterval(f, lk, best, math.Log10(kLo), math.Log10(kHi))
	return math.Pow(10, lk), best, interval, true
}

// InvertLayered estimates two layer permeabilities (inlet half, outlet half).
//
// The cost surface in (log k1, log k2) has a narrow global basin and broad
// secondary basins, so a cold start in the middle of the range is liable to
// converge to a compromise solution (both layers drifting toward a mid value).
// We therefore first coarsely scan the full range on a logarithmic grid to
// locate the global basin, then refine with alternating coordinate descent
// seeded from that basin.
func InvertLayered(fg FluidGeom, bc BoundaryCond, cfg SimConfig, obs []Observation, kLo, kHi float64) (k1, k2, sse float64, interval [2]float64, converged bool) {
	lo, hi := math.Log10(kLo), math.Log10(kHi)
	build := func(l1, l2 float64) Medium {
		return Medium{Layers: []Layer{
			{Permeability: math.Pow(10, l1), ThicknessFrac: 0.5},
			{Permeability: math.Pow(10, l2), ThicknessFrac: 0.5},
		}}
	}

	// Coarse global scan on a logarithmic grid.
	const gridN = 40
	bl1, bl2 := lo, lo
	best := math.Inf(1)
	for i := 0; i < gridN; i++ {
		l1 := lo + (hi-lo)*float64(i)/float64(gridN-1)
		for j := 0; j < gridN; j++ {
			l2 := lo + (hi-lo)*float64(j)/float64(gridN-1)
			if c := cost(build(l1, l2), fg, bc, cfg, obs); c < best {
				best, bl1, bl2 = c, l1, l2
			}
		}
	}

	// Refine with alternating coordinate descent from the located basin.
	bk1, bk2 := bl1, bl2
	for round := 0; round < 8; round++ {
		fK1 := func(l1 float64) float64 { return cost(build(l1, bk2), fg, bc, cfg, obs) }
		bk1, _ = goldenSection(fK1, lo, hi, 40)
		fK2 := func(l2 float64) float64 { return cost(build(bk1, l2), fg, bc, cfg, obs) }
		bk2, _ = goldenSection(fK2, lo, hi, 40)
	}
	m := build(bk1, bk2)
	sse = cost(m, fg, bc, cfg, obs)
	fFull := func(l1 float64) float64 { return cost(build(l1, bk2), fg, bc, cfg, obs) }
	interval = identifiabilityInterval(fFull, bk1, sse, lo, hi)
	return math.Pow(10, bk1), math.Pow(10, bk2), sse, interval, true
}

// ModelResult summarises one candidate model's inversion outcome.
type ModelResult struct {
	Form          string
	SSE           float64
	NormalizedSSE float64
	PrimaryK      float64
	Interval      [2]float64
	Converged     bool
	// Fits reports whether the model explains the observations within tolerance
	// (combining a relative and an absolute floor so zero-energy data is
	// judged by absolute residual).
	Fits bool
}

// FitsWell judges whether a residual sum-of-squares explains the observations.
// It blends a relative tolerance on the normalised energy with an absolute
// floor so that near-zero observations are not mis-scored.
func FitsWell(sse float64, count int, energy, relTol, absTol float64) bool {
	if count <= 0 {
		return false
	}
	thresh := relTol*float64(count)*energy + absTol*float64(count)
	return sse <= thresh
}

// Comparison aggregates candidate-model results and judges identifiability.
type Comparison struct {
	Results      []ModelResult
	Identifiable bool
	Reason       string
}

// CompareModels decides whether the candidate models are distinguishable.
// Structural non-identifiability arises when two or more candidate structures
// both fit the data (comparable residuals): the observations cannot tell them
// apart, so more surveys are required. If exactly one candidate fits, the
// correct boundary structure is identifiable.
func CompareModels(results []ModelResult, fitTol float64) Comparison {
	nFit := 0
	for _, r := range results {
		if r.Converged && r.Fits {
			nFit++
		}
	}
	if len(results) >= 2 && nFit >= 2 {
		return Comparison{
			Results:      results,
			Identifiable: false,
			Reason:       "multiple candidate models fit the observations equally well; the boundary structure is non-identifiable (add more observations)",
		}
	}
	best := results[0]
	for _, r := range results {
		if r.NormalizedSSE < best.NormalizedSSE {
			best = r
		}
	}
	return Comparison{
		Results:      results,
		Identifiable: true,
		Reason:       "only the " + best.Form + " candidate fits within tolerance; the boundary structure is identifiable",
	}
}

// MeanObservationEnergy returns the mean squared observation magnitude, used to
// normalise the cost when judging fit quality.
func MeanObservationEnergy(obs []Observation) float64 {
	if len(obs) == 0 {
		return 1
	}
	s := 0.0
	for _, o := range obs {
		s += o.Value * o.Value
	}
	return s / float64(len(obs))
}

// SortObservations orders observations by time for stable downstream use.
func SortObservations(obs []Observation) {
	sort.Slice(obs, func(i, j int) bool {
		if obs[i].T == obs[j].T {
			return obs[i].XFrac < obs[j].XFrac
		}
		return obs[i].T > obs[j].T
	})
}
