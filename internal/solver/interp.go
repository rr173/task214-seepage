// Package solver implements a one-dimensional transient Darcy pressure-diffusion
// forward model for a porous column together with boundary-parameter inversion
// and candidate-model identifiability analysis.
package solver

// PiecewiseLinear builds a linear-interpolation function from ascending sample
// points (ts[i], vs[i]). Values outside the sampled window are clamped to the
// nearest end point, matching how uploaded boundary sequences are replayed
// during inversion. Sharing this single implementation between the data
// generator and the boundary reconstruction guarantees the two stay in exact
// agreement.
func PiecewiseLinear(ts, vs []float64) func(float64) float64 {
	return func(t float64) float64 {
		if len(ts) == 0 {
			return 0
		}
		if len(vs) == 0 {
			return 0
		}
		if t <= ts[0] {
			return vs[0]
		}
		if t >= ts[len(ts)-1] {
			return vs[len(vs)-1]
		}
		for i := 1; i < len(ts); i++ {
			if t <= ts[i] {
				span := ts[i] - ts[i-1]
				if span == 0 {
					return vs[i]
				}
				frac := (t - ts[i-1]) / span
				return vs[i-1] + frac*(vs[i]-vs[i-1])
			}
		}
		return vs[len(vs)-1]
	}
}
