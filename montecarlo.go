package banzhaf

import (
	"math"
	"math/rand"
)

// MonteCarlo estimates the absolute Banzhaf index eta_i / 2^(n-1) by sampling
// random coalitions. A single random membership vector yields an unbiased swing
// indicator for every player at once, so cost is O(samples * n) total. By
// Hoeffding, samples = ceil(ln(2n/delta) / (2*eps^2)) give a uniform eps-bound at
// confidence 1-delta on the absolute index, independent of W.
//
// Sampling is a last resort: it is accurate for absolute indices or when some
// players are genuinely pivotal, but on flat low-pivotality electorates the
// normalized index is noisy. Prefer HighN.
func MonteCarlo(weights []float64, quota float64, samples int, seed int64) []float64 {
	n := len(weights)
	idx := make([]float64, n)
	if n == 0 || samples <= 0 {
		return idx
	}
	rng := rand.New(rand.NewSource(seed))
	crit := make([]float64, n)
	for s := 0; s < samples; s++ {
		var total float64
		in := make([]bool, n)
		for i, w := range weights {
			if rng.Intn(2) == 1 {
				in[i] = true
				total += w
			}
		}
		for i, w := range weights {
			others := total
			if in[i] {
				others -= w
			}
			if others >= quota-w && others <= quota-1 {
				crit[i]++
			}
		}
	}
	for i := range crit {
		idx[i] = crit[i] / float64(samples)
	}
	return idx
}

// SamplesFor returns the Hoeffding sample count for a uniform additive eps-bound
// on the absolute index across all n players at confidence 1-delta.
func SamplesFor(n int, eps, delta float64) int {
	return int(math.Ceil(math.Log(2*float64(n)/delta) / (2 * eps * eps)))
}
