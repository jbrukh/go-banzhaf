package banzhaf

import "math"

// NormalApprox returns the absolute Banzhaf index eta_i / 2^(n-1) for every
// player via an O(n) Gaussian (central-limit) approximation, independent of the
// total weight W. The weight of a random subset of the other players is a sum of
// independent {0, w_j} variables, approximately Normal(mu_i, var_i); the swing
// probability follows from the normal CDF over the width-w_i window at the quota.
//
// Accuracy improves with n and as no single weight dominates the variance. For
// distributions with a few whales prefer HighN, which handles them exactly.
func NormalApprox(weights []float64, quota float64) []float64 {
	n := len(weights)
	idx := make([]float64, n)
	if n == 0 {
		return idx
	}

	var s1, s2 float64 // sum and sum-of-squares of all weights
	for _, w := range weights {
		s1 += w
		s2 += w * w
	}

	for i, wi := range weights {
		if wi == 0 {
			continue
		}
		mu := 0.5 * (s1 - wi)
		varr := 0.25 * (s2 - wi*wi)
		if varr <= 0 {
			if wi >= quota {
				idx[i] = 1
			}
			continue
		}
		sd := math.Sqrt(varr)
		hi := (quota - 0.5 - mu) / sd
		lo := (quota - wi - 0.5 - mu) / sd
		idx[i] = normCDF(hi) - normCDF(lo)
	}
	return idx
}
