package banzhaf

import (
	"fmt"
	"math/big"
)

// Incremental maintains the generating polynomial F = prod_j (1 + x^{w_j}) so a
// single balance change is an O(W) edit (multiply or divide one factor) instead
// of an O(n*W) rebuild. Because changing balances also moves the quota, F is kept
// untruncated and remains valid for any quota the current supply implies.
//
// All arithmetic is exact big integers; intended for small-weight (governance)
// games and live dashboards, not raw on-chain magnitudes.
type Incremental struct {
	counts map[uint64]int // weight -> number of holders
	total  uint64         // total supply
	f      []*big.Int     // coefficients of prod (1 + x^{w_j})
}

// NewIncremental builds the state from an initial set of weights.
func NewIncremental(weights []uint64) *Incremental {
	in := &Incremental{counts: map[uint64]int{}, f: []*big.Int{big.NewInt(1)}}
	for _, w := range weights {
		in.Add(w)
	}
	return in
}

// Add inserts one holder of weight w in O(W).
func (in *Incremental) Add(w uint64) {
	in.counts[w]++
	in.total += w
	if w == 0 {
		for k := range in.f {
			in.f[k].Lsh(in.f[k], 1)
		}
		return
	}
	f := in.f
	g := make([]*big.Int, len(f)+int(w))
	for k := range g {
		g[k] = new(big.Int)
	}
	for k := range f {
		g[k].Add(g[k], f[k])
		g[k+int(w)].Add(g[k+int(w)], f[k])
	}
	in.f = g
}

// Remove deletes one holder of weight w (which must exist) in O(W).
func (in *Incremental) Remove(w uint64) error {
	if in.counts[w] <= 0 {
		return fmt.Errorf("no holder with weight %d to remove", w)
	}
	in.counts[w]--
	if in.counts[w] == 0 {
		delete(in.counts, w)
	}
	in.total -= w
	if w == 0 {
		for k := range in.f {
			in.f[k].Rsh(in.f[k], 1)
		}
		return nil
	}
	f := in.f
	m := len(f) - int(w)
	g := make([]*big.Int, m)
	for k := 0; k < m; k++ {
		g[k] = new(big.Int)
		if k < int(w) {
			g[k].Set(f[k])
		} else {
			g[k].Sub(f[k], g[k-int(w)])
		}
	}
	in.f = g
	return nil
}

// Change moves one holder from w_old to w_new in O(W).
func (in *Incremental) Change(wOld, wNew uint64) error {
	if err := in.Remove(wOld); err != nil {
		return err
	}
	in.Add(wNew)
	return nil
}

// Total returns the current total supply.
func (in *Incremental) Total() uint64 { return in.total }

// Scores returns the raw swing counts eta keyed by distinct weight, O(d*W).
func (in *Incremental) Scores(quota uint64) (map[uint64]*big.Int, error) {
	if quota < 1 || quota > in.total {
		return nil, fmt.Errorf("quota %d out of range [1,%d]", quota, in.total)
	}
	cap := quota - 1
	out := make(map[uint64]*big.Int, len(in.counts))
	g := make([]*big.Int, cap+1)
	for k := range g {
		g[k] = new(big.Int)
	}
	for w := range in.counts {
		if w == 0 {
			out[0] = new(big.Int)
			continue
		}
		for k := uint64(0); k <= cap; k++ {
			fk := zero
			if int(k) < len(in.f) {
				fk = in.f[k]
			}
			if k < w {
				g[k].Set(fk)
			} else {
				g[k].Sub(fk, g[k-w])
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
		out[w] = sum
	}
	return out, nil
}

// ScoresList returns eta for an explicit weight list (replicated by symmetry).
func (in *Incremental) ScoresList(weights []uint64, quota uint64) ([]*big.Int, error) {
	byW, err := in.Scores(quota)
	if err != nil {
		return nil, err
	}
	out := make([]*big.Int, len(weights))
	for i, w := range weights {
		out[i] = new(big.Int).Set(byW[w])
	}
	return out, nil
}
