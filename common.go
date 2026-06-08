package banzhaf

import (
	"math"
	"math/big"
)

// invSqrt2 = 1/sqrt(2), used to express the normal CDF via the error function.
var invSqrt2 = 1.0 / math.Sqrt2

// normPDF is the standard normal probability density at z.
func normPDF(z float64) float64 {
	return math.Exp(-0.5*z*z) / math.Sqrt(2*math.Pi)
}

// normCDF is the standard normal cumulative distribution at z, computed from
// the complementary error function for good tail behavior.
func normCDF(z float64) float64 {
	return 0.5 * math.Erfc(-z*invSqrt2)
}

// he3 and he4 are the (probabilists') Hermite polynomials used by the Edgeworth
// correction in the high-n method.
func he3(z float64) float64 { return z * (z*z - 3) }
func he4(z float64) float64 { z2 := z * z; return z2*(z2-6) + 3 }

// Normalize scales raw scores so they sum to one (the normalized power index).
// A zero total yields all-zero output.
func Normalize(scores []float64) []float64 {
	var total float64
	for _, s := range scores {
		total += s
	}
	out := make([]float64, len(scores))
	if total == 0 {
		return out
	}
	for i, s := range scores {
		out[i] = s / total
	}
	return out
}

// normalizeBig turns exact big.Int swing counts into a normalized float index.
func normalizeBig(eta []*big.Int) []float64 {
	total := new(big.Int)
	for _, e := range eta {
		total.Add(total, e)
	}
	out := make([]float64, len(eta))
	if total.Sign() == 0 {
		return out
	}
	d := new(big.Float).SetInt(total)
	for i, e := range eta {
		q := new(big.Float).Quo(new(big.Float).SetInt(e), d)
		out[i], _ = q.Float64()
	}
	return out
}

// absoluteBig turns exact swing counts into the absolute index eta_i / 2^(n-1).
func absoluteBig(eta []*big.Int) []float64 {
	n := len(eta)
	out := make([]float64, n)
	if n == 0 {
		return out
	}
	denom := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(2), big.NewInt(int64(n-1)), nil))
	for i, e := range eta {
		q := new(big.Float).Quo(new(big.Float).SetInt(e), denom)
		out[i], _ = q.Float64()
	}
	return out
}

// sumU64 returns the total weight; ok is false on uint64 overflow.
func sumU64(weights []uint64) (total uint64, ok bool) {
	for _, w := range weights {
		next := total + w
		if next < total {
			return 0, false
		}
		total = next
	}
	return total, true
}

// distinctCount returns the number of distinct weights.
func distinctCount(weights []float64) int {
	seen := make(map[float64]struct{}, len(weights))
	for _, w := range weights {
		seen[w] = struct{}{}
	}
	return len(seen)
}
