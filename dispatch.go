package banzhaf

import (
	"fmt"
	"math"
)

// Strategy names the algorithm the dispatcher selected.
type Strategy string

const (
	// StrategyExact: grouped generating-function DP, exact, O((n+d)*W). Chosen
	// when the weights are integral and the total weight is small enough to
	// afford an array of that size.
	StrategyExact Strategy = "exact-grouped"
	// StrategyHighN: head/tail density approximation, O(n+2^k), W-independent.
	// Chosen for high n and/or large raw supplies where no weight-axis array is
	// affordable.
	StrategyHighN Strategy = "highn"
)

// Options controls the dispatcher.
type Options struct {
	// MaxArrayWeight is the largest total weight for which the exact array-based
	// method is considered affordable. Above it (or for non-integer weights) the
	// approximation is used. Zero selects the default (2,000,000).
	MaxArrayWeight uint64
	// ForceExact errors instead of approximating if the exact method is not
	// affordable, rather than silently switching strategy.
	ForceExact bool
	// HighN tunes the approximation when it is selected.
	HighN HighNOptions
}

// Result is the outcome of Compute.
type Result struct {
	// Index is the normalized Banzhaf power index (sums to 1).
	Index []float64
	// Absolute is the absolute index eta_i / 2^(n-1) (a probability in [0,1]).
	Absolute []float64
	// Strategy is the algorithm that was chosen.
	Strategy Strategy
	// Exact reports whether the result is exact (true) or approximate (false).
	Exact bool
}

const defaultMaxArrayWeight = 2_000_000

// Compute inspects the data and returns the Banzhaf power index using the best
// available strategy: the exact grouped DP when the weights are integral and the
// total weight is small enough to afford an array, otherwise the W-independent
// high-n approximation. This is the recommended entry point.
//
// Weights are float64 so that raw on-chain balances (which overflow uint64 and
// any practical array) are accepted directly. quota is the winning threshold;
// pass 0 to default to a strict majority (floor(W/2)+1).
func Compute(weights []float64, quota float64, opts Options) (Result, error) {
	n := len(weights)
	if n == 0 {
		return Result{Index: []float64{}, Absolute: []float64{}, Strategy: StrategyHighN}, nil
	}
	maxW := opts.MaxArrayWeight
	if maxW == 0 {
		maxW = defaultMaxArrayWeight
	}

	var total float64
	for _, w := range weights {
		if w < 0 {
			return Result{}, fmt.Errorf("negative weight %g", w)
		}
		total += w
	}
	if quota == 0 {
		quota = math.Floor(total/2) + 1
	}

	// Is the exact array-based method affordable? Requires integral weights and a
	// total weight that fits both uint64 and the array budget.
	exactAffordable := integral(weights) && integral([]float64{quota}) &&
		total <= float64(maxW) && total <= math.MaxInt64

	if !exactAffordable {
		if opts.ForceExact {
			return Result{}, fmt.Errorf(
				"exact method not affordable (total weight %.0f, non-integer, or above MaxArrayWeight %d)",
				total, maxW)
		}
		abs := HighN(weights, quota, opts.HighN)
		return Result{Index: Normalize(abs), Absolute: abs, Strategy: StrategyHighN, Exact: false}, nil
	}

	// Exact path.
	uw := make([]uint64, n)
	for i, w := range weights {
		uw[i] = uint64(w)
	}
	eta, err := Exact(uw, uint64(quota))
	if err != nil {
		return Result{}, err
	}
	return Result{
		Index:    normalizeBig(eta),
		Absolute: absoluteBig(eta),
		Strategy: StrategyExact,
		Exact:    true,
	}, nil
}

// integral reports whether every value is a non-negative whole number.
func integral(xs []float64) bool {
	for _, x := range xs {
		if x != math.Trunc(x) || x < 0 {
			return false
		}
	}
	return true
}
