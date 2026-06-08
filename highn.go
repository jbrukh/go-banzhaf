package banzhaf

import (
	"math"
	"sort"
)

// HighNOptions tunes the high-n approximation.
type HighNOptions struct {
	// MaxHead caps the exact "head" at k holders; cost is O(n + k*2^k). On
	// scale-free (power-law) distributions the head fills to this cap and 2^k
	// dominates the runtime. 14 is the measured sweet spot for token snapshots
	// (~10 ms at 100k holders, ~0.02% error on top holders). Raise for sub-0.01%.
	MaxHead int
	// Edgeworth adds the tail's kurtosis correction. The fair-coin tail is exactly
	// symmetric (zero skew), so kurtosis is the leading term; in practice it is
	// marginal at this accuracy, so it is off by default.
	Edgeworth bool
}

// DefaultHighN is the recommended configuration.
var DefaultHighN = HighNOptions{MaxHead: 14, Edgeworth: false}

// HighN returns the absolute Banzhaf index eta_i / 2^(n-1) via the head/tail
// density method: the top-k holders ("head") are handled exactly by enumerating
// their 2^k subset sums, while the smooth "tail" is treated as a Gaussian by the
// central limit theorem. Runs in O(n + k*2^k) and is completely independent of
// the total weight W, so it is the method for high n with large raw on-chain
// supplies. Pass HighNOptions{} to use the defaults.
//
// A small holder swings only when the rest land in a width-w_i window at the
// quota, so its index is w_i * a(q) where a is the aggregate density at the quota
// (this linear form is second-order accurate via an exact cancellation between
// window curvature and the leave-one-out density shift). Each head holder is
// evaluated exactly by leave-one-out, all sharing one head enumeration.
func HighN(weights []float64, quota float64, opts HighNOptions) []float64 {
	if opts.MaxHead == 0 {
		opts.MaxHead = DefaultHighN.MaxHead
	}
	n := len(weights)
	idx := make([]float64, n)
	switch n {
	case 0:
		return idx
	case 1:
		if weights[0] >= quota {
			idx[0] = 1
		}
		return idx
	}

	headPos := chooseHead(weights, opts.MaxHead)
	isHead := make([]bool, n)
	for _, p := range headPos {
		isHead[p] = true
	}
	headW := make([]float64, len(headPos))
	for i, p := range headPos {
		headW[i] = weights[p]
	}

	// tail moments
	var muT, varT float64
	for i, w := range weights {
		if !isHead[i] {
			muT += 0.5 * w
			varT += 0.25 * w * w
		}
	}

	if varT <= 0 {
		// degenerate tail: exact enumeration over the whole (tiny) set
		for i := range weights {
			others := make([]float64, 0, n-1)
			others = append(others, weights[:i]...)
			others = append(others, weights[i+1:]...)
			sums := subsetSums(others)
			var hit float64
			for _, s := range sums {
				if s >= quota-weights[i] && s <= quota-1 {
					hit++
				}
			}
			idx[i] = hit / float64(len(sums))
		}
		return idx
	}
	sdT := math.Sqrt(varT)

	// kurtosis correction for the tail (zero skew by symmetry)
	g2 := 0.0
	if opts.Edgeworth {
		var k2, k4 float64
		for i, w := range weights {
			if !isHead[i] {
				k2 += 0.25 * w * w
				k4 += -0.125 * w * w * w * w
			}
		}
		if k2 > 0 {
			g2 = k4 / (k2 * k2)
		}
	}

	headSums := subsetSums(headW) // 2^k atoms, equal weight; bit p == head[p] present

	// --- tail holders: index = w_i * a(q) ---------------------------------
	// a(q): Edgeworth-corrected aggregate subset-sum density at the quota.
	var a0 float64
	for _, h := range headSums {
		z := (quota - 0.5 - h - muT) / sdT
		a0 += normPDF(z) * (1 + (g2/24)*he4(z))
	}
	a0 /= float64(len(headSums)) * sdT
	for i, w := range weights {
		if !isHead[i] {
			idx[i] = w * a0
		}
	}

	// --- head holders: exact leave-one-out, batched -----------------------
	// subsets of head\{pos} are the indices with bit pos cleared.
	cdf := func(z float64) float64 {
		return normCDF(z) - (g2/24)*normPDF(z)*he3(z)
	}
	for pos, gi := range headPos {
		wj := weights[gi]
		var swing float64
		count := 0
		for i, h := range headSums {
			if (i>>uint(pos))&1 == 1 {
				continue // pos present -> not a subset of head\{pos}
			}
			zHi := (quota - 0.5 - h - muT) / sdT
			zLo := (quota - wj - 0.5 - h - muT) / sdT
			swing += cdf(zHi) - cdf(zLo)
			count++
		}
		idx[gi] = swing / float64(count)
	}
	return idx
}

// chooseHead returns the positions of the head: the largest holders, grown until
// the largest remaining holder is small versus the tail's standard deviation,
// capped at maxHead.
func chooseHead(weights []float64, maxHead int) []int {
	n := len(weights)
	order := make([]int, n)
	for i := range order {
		order[i] = i
	}
	sort.Slice(order, func(a, b int) bool { return weights[order[a]] > weights[order[b]] })

	limit := maxHead
	if limit > n {
		limit = n
	}
	k := 0
	for k < limit {
		k++
		// standard deviation of the tail beyond k
		var s2 float64
		for _, p := range order[k:] {
			s2 += weights[p] * weights[p]
		}
		if len(order) == k {
			break
		}
		sdTail := math.Sqrt(0.25 * s2)
		if sdTail == 0 {
			sdTail = 1
		}
		if weights[order[k]] <= 0.15*sdTail {
			break
		}
	}
	return order[:k]
}

// subsetSums returns all 2^len(weights) subset sums; element i is the sum of the
// subset whose bit pattern is i (bit p == weights[p] included).
func subsetSums(weights []float64) []float64 {
	sums := make([]float64, 1, 1<<uint(len(weights)))
	for _, w := range weights {
		m := len(sums)
		for i := 0; i < m; i++ {
			sums = append(sums, sums[i]+w)
		}
	}
	return sums
}
