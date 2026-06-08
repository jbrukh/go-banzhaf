// Walkthrough: compute the Banzhaf power index on a token-holder snapshot, both
// exactly and approximately, and measure runtime/cost.
package main

import (
	"fmt"
	"math/big"
	"math/rand"
	"sort"
	"time"

	banzhaf "github.com/jbrukh/go-banzhaf"
)

// normBig turns exact big.Int swing counts into a normalized float index.
func normBig(eta []*big.Int) []float64 {
	total := new(big.Int)
	for _, e := range eta {
		total.Add(total, e)
	}
	out := make([]float64, len(eta))
	d := new(big.Float).SetInt(total)
	for i, e := range eta {
		f, _ := new(big.Float).Quo(new(big.Float).SetInt(e), d).Float64()
		out[i] = f
	}
	return out
}

func main() {
	// -----------------------------------------------------------------------
	// A dataset: a small DAO snapshot with integer "vote weights" so that the
	// exact method is tractable and we can compare the approximation to truth.
	//   5 whales, ~12 mid holders, ~300 small holders.
	// -----------------------------------------------------------------------
	rng := rand.New(rand.NewSource(42))
	var w []float64
	w = append(w, 40, 38, 35, 30, 28) // whales
	for i := 0; i < 12; i++ {
		w = append(w, float64(5+rng.Intn(11))) // mids 5..15
	}
	for i := 0; i < 300; i++ {
		w = append(w, float64(1+rng.Intn(3))) // tail 1..3
	}

	var total float64
	for _, x := range w {
		total += x
	}
	quota := total/2 + 1 // strict majority
	fmt.Printf("Dataset: n=%d holders, total weight W=%.0f, quota=%.0f (strict majority)\n",
		len(w), total, quota)
	fmt.Printf("Top weights: %v ... long tail of 1-3\n\n", w[:5])

	// -----------------------------------------------------------------------
	// 1) EXACT — grouped generating-function DP, O((n+d)*W), big integers.
	// -----------------------------------------------------------------------
	uw := make([]uint64, len(w))
	for i, x := range w {
		uw[i] = uint64(x)
	}
	t0 := time.Now()
	eta, err := banzhaf.Exact(uw, uint64(quota))
	if err != nil {
		panic(err)
	}
	exactNorm := normBig(eta)
	exactTime := time.Since(t0)

	// -----------------------------------------------------------------------
	// 2) APPROXIMATE — HighN head/tail density, O(n + k*2^k), W-independent.
	// -----------------------------------------------------------------------
	t1 := time.Now()
	approxNorm := banzhaf.Normalize(banzhaf.HighN(w, quota, banzhaf.HighNOptions{}))
	approxTime := time.Since(t1)

	// rank holders by exact power and show the top few, both ways
	idx := make([]int, len(w))
	for i := range idx {
		idx[i] = i
	}
	sort.Slice(idx, func(a, b int) bool { return exactNorm[idx[a]] > exactNorm[idx[b]] })

	fmt.Println("Power index of the most powerful holders:")
	fmt.Printf("  %-8s %-8s %-14s %-14s %-10s\n", "holder", "weight", "exact β", "approx β", "rel.err")
	var maxRel float64
	for _, r := range idx[:8] {
		rel := abs(approxNorm[r]-exactNorm[r]) / exactNorm[r]
		if rel > maxRel {
			maxRel = rel
		}
		fmt.Printf("  #%-7d %-8.0f %-14.8f %-14.8f %-10.1e\n",
			r, w[r], exactNorm[r], approxNorm[r], rel)
	}
	fmt.Printf("\nMax relative error on the top-8 holders: %.2e\n", maxRel)
	fmt.Printf("Runtime:  exact = %v    approx = %v\n",
		exactTime.Round(time.Microsecond), approxTime.Round(time.Microsecond))
	if exactTime < approxTime {
		fmt.Printf("  -> at this small W, EXACT is %.1fx faster: HighN's fixed 2^k head\n"+
			"     overhead isn't worth paying when an array of size W is cheap.\n\n",
			float64(approxTime)/float64(exactTime))
	} else {
		fmt.Printf("  -> approx is %.1fx faster\n\n", float64(exactTime)/float64(approxTime))
	}

	// -----------------------------------------------------------------------
	// 3) The dispatcher picks the strategy automatically.
	// -----------------------------------------------------------------------
	res, _ := banzhaf.Compute(w, quota, banzhaf.Options{})
	fmt.Printf("Compute() auto-selected strategy: %q (exact=%v)\n\n", res.Strategy, res.Exact)

	// -----------------------------------------------------------------------
	// 4) REALITY: scale to raw on-chain units and 100k holders.
	//    Now W ~ 10^25 — the exact method is impossible (it would allocate an
	//    array of size W). Only the approximation runs.
	// -----------------------------------------------------------------------
	fmt.Println("--- Same shape, but raw on-chain scale: balances x 10^18, 100k holders ---")
	big := make([]float64, 0, 100_000)
	big = append(big, 40e18, 38e18, 35e18, 30e18, 28e18)
	for i := 0; i < 12; i++ {
		big = append(big, float64(5+rng.Intn(11))*1e18)
	}
	for i := 0; i < 100_000-17; i++ {
		big = append(big, float64(1+rng.Intn(3))*1e18)
	}
	var bigTotal float64
	for _, x := range big {
		bigTotal += x
	}
	bq := bigTotal/2 + 1

	t2 := time.Now()
	bigRes, _ := banzhaf.Compute(big, bq, banzhaf.Options{})
	bigTime := time.Since(t2)
	// top holder of the scaled set
	topVal := 0.0
	for _, v := range bigRes.Index {
		if v > topVal {
			topVal = v
		}
	}
	fmt.Printf("n=%d, W=%.2e  ->  strategy=%q, exact=%v, time=%v\n",
		len(big), bigTotal, bigRes.Strategy, bigRes.Exact, bigTime.Round(time.Microsecond))
	fmt.Printf("top holder normalized power = %.6f\n", topVal)

	// exact is not even an option here:
	if _, err := banzhaf.Compute(big, bq, banzhaf.Options{ForceExact: true}); err != nil {
		fmt.Printf("ForceExact at this scale: %v\n", err)
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
