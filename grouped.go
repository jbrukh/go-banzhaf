package banzhaf

import (
	"fmt"
	"math/big"
)

// Exact returns the raw Banzhaf swing counts eta_i for a weighted voting game,
// using the grouped generating-function method. It is exact (arbitrary-precision
// big integers) and runs in O((n+d)*W) time and O(W) space, where n is the
// number of players, d the number of distinct weights, and W the total weight.
//
// Players that share a weight have identical scores by symmetry, so the per-
// weight divide-out and window-sum is done once per distinct weight and
// replicated. The polynomial is truncated at the quota, since no swing window
// ever reads above it.
//
// quota must satisfy 1 <= quota <= total weight. Use Normalize / absolute helpers
// to turn the counts into a power index.
func Exact(weights []uint64, quota uint64) ([]*big.Int, error) {
	n := len(weights)
	total, ok := sumU64(weights)
	if !ok {
		return nil, fmt.Errorf("total weight overflows uint64; use the approximate methods")
	}
	if quota < 1 || quota > total {
		return nil, fmt.Errorf("quota %d out of range [1,%d]", quota, total)
	}

	cap := quota - 1
	// f[k] = number of coalitions of total weight k, truncated to [0,cap].
	f := make([]*big.Int, cap+1)
	for k := range f {
		f[k] = new(big.Int)
	}
	f[0].SetUint64(1)
	for _, w := range weights {
		if w == 0 {
			// a zero-weight player doubles every coalition count
			for k := range f {
				f[k].Lsh(f[k], 1)
			}
			continue
		}
		if w > cap {
			continue
		}
		// multiply by (1 + x^w): f[k] += f[k-w], descending to stay in place
		for k := cap; k >= w; k-- {
			f[k].Add(f[k], f[k-w])
		}
	}

	// one divide-out + window sum per distinct weight
	etaByW := make(map[uint64]*big.Int)
	g := make([]*big.Int, cap+1)
	for k := range g {
		g[k] = new(big.Int)
	}
	for _, w := range weights {
		if w == 0 {
			etaByW[0] = new(big.Int) // never critical
			continue
		}
		if _, done := etaByW[w]; done {
			continue
		}
		// divide out (1 + x^w): g[k] = f[k] - g[k-w]
		for k := uint64(0); k <= cap; k++ {
			if k < w {
				g[k].Set(f[k])
			} else {
				g[k].Sub(f[k], g[k-w])
			}
		}
		lo := uint64(0)
		if quota > w {
			lo = quota - w
		}
		sum := new(big.Int)
		for k := lo; k <= cap; k++ {
			sum.Add(sum, g[k])
		}
		etaByW[w] = sum
	}

	eta := make([]*big.Int, n)
	for i, w := range weights {
		eta[i] = new(big.Int).Set(etaByW[w])
	}
	return eta, nil
}
