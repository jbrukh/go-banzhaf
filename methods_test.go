package banzhaf

import (
	"math"
	"math/big"
	"math/rand"
	"testing"
)

// naive is an independent O(2^n * n) brute force used only as ground truth.
func naive(weights []uint64, quota uint64) []*big.Int {
	n := len(weights)
	eta := make([]*big.Int, n)
	for i := range eta {
		eta[i] = new(big.Int)
	}
	for mask := uint64(0); mask < (1 << uint(n)); mask++ {
		var total uint64
		for i := 0; i < n; i++ {
			if mask&(1<<uint(i)) != 0 {
				total += weights[i]
			}
		}
		if total < quota {
			continue
		}
		for i := 0; i < n; i++ {
			if mask&(1<<uint(i)) != 0 && total-weights[i] < quota {
				eta[i].Add(eta[i], big.NewInt(1))
			}
		}
	}
	return eta
}

func bigEqual(a, b []*big.Int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Cmp(b[i]) != 0 {
			return false
		}
	}
	return true
}

// TestExactVsBruteForce cross-checks the grouped DP against brute force over many
// random games and several quotas.
func TestExactVsBruteForce(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	for iter := 0; iter < 300; iter++ {
		n := 1 + rng.Intn(12)
		w := make([]uint64, n)
		var total uint64
		for i := range w {
			w[i] = uint64(rng.Intn(9))
			total += w[i]
		}
		if total == 0 {
			continue
		}
		for _, q := range []uint64{total/2 + 1, total, (total + 2) / 3} {
			if q < 1 || q > total {
				continue
			}
			got, err := Exact(w, q)
			if err != nil {
				t.Fatalf("Exact: %v", err)
			}
			if want := naive(w, q); !bigEqual(got, want) {
				t.Fatalf("mismatch w=%v q=%d: got %v want %v", w, q, got, want)
			}
		}
	}
}

// TestIncrementalMatchesExact applies a random edit sequence and re-checks against
// a from-scratch exact computation each step.
func TestIncrementalMatchesExact(t *testing.T) {
	rng := rand.New(rand.NewSource(2))
	weights := []uint64{}
	for i := 0; i < 12; i++ {
		weights = append(weights, uint64(1+rng.Intn(5)))
	}
	game := NewIncremental(weights)

	for step := 0; step < 40; step++ {
		switch rng.Intn(3) {
		case 0:
			w := uint64(1 + rng.Intn(5))
			weights = append(weights, w)
			game.Add(w)
		case 1:
			if len(weights) > 1 {
				j := rng.Intn(len(weights))
				w := weights[j]
				weights = append(weights[:j], weights[j+1:]...)
				if err := game.Remove(w); err != nil {
					t.Fatal(err)
				}
			}
		case 2:
			if len(weights) > 0 {
				j := rng.Intn(len(weights))
				wOld, wNew := weights[j], uint64(1+rng.Intn(5))
				weights[j] = wNew
				if err := game.Change(wOld, wNew); err != nil {
					t.Fatal(err)
				}
			}
		}
		total := game.Total()
		q := total/2 + 1
		got, err := game.ScoresList(weights, q)
		if err != nil {
			t.Fatal(err)
		}
		want, err := Exact(weights, q)
		if err != nil {
			t.Fatal(err)
		}
		if !bigEqual(got, want) {
			t.Fatalf("step %d: incremental != exact for %v", step, weights)
		}
	}
}

// TestMonteCarloWithinBound checks the estimator respects its Hoeffding bound.
func TestMonteCarloWithinBound(t *testing.T) {
	rng := rand.New(rand.NewSource(3))
	w := make([]float64, 25)
	var uw []uint64
	for i := range w {
		v := 1 + rng.Intn(5)
		w[i] = float64(v)
		uw = append(uw, uint64(v))
	}
	var total uint64
	for _, x := range uw {
		total += x
	}
	q := total/2 + 1
	eta, _ := Exact(uw, q)
	exactAbs := absoluteBig(eta)

	eps, delta := 0.01, 0.01
	s := SamplesFor(len(w), eps, delta)
	got := MonteCarlo(w, float64(q), s, 1)
	if d := maxAbsDiff(got, exactAbs); d > 1.5*eps {
		t.Errorf("monte carlo exceeded bound: %.2e > %.2e", d, 1.5*eps)
	}
}

// TestDispatcherSelectsStrategy verifies the data-driven strategy choice.
func TestDispatcherSelectsStrategy(t *testing.T) {
	// small integer weights -> exact
	r, err := Compute([]float64{2, 2, 1}, 0, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if r.Strategy != StrategyExact || !r.Exact {
		t.Errorf("small game should be exact, got %s", r.Strategy)
	}

	// huge raw on-chain weights (overflow uint64 and any array) -> high-n
	big := []float64{6e23, 5e23, 1e22, 3e22, 2e22, 8e21, 4e21}
	r2, err := Compute(big, 0, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if r2.Strategy != StrategyHighN || r2.Exact {
		t.Errorf("huge-supply game should be high-n, got %s", r2.Strategy)
	}
	if s := sum(r2.Index); math.Abs(s-1) > 1e-9 {
		t.Errorf("normalized index should sum to 1, got %g", s)
	}

	// ForceExact on an unaffordable game errors instead of silently approximating
	if _, err := Compute(big, 0, Options{ForceExact: true}); err == nil {
		t.Error("ForceExact on huge supply should error")
	}
}

// TestDispatcherExactMatchesWeightAxis confirms the dispatcher's exact path equals
// the standalone Exact function.
func TestDispatcherExactMatchesWeightAxis(t *testing.T) {
	w := []float64{5, 4, 3, 2, 2, 1, 1}
	var total uint64 = 18
	q := total/2 + 1
	r, _ := Compute(w, float64(q), Options{})
	uw := []uint64{5, 4, 3, 2, 2, 1, 1}
	eta, _ := Exact(uw, q)
	if maxAbsDiff(r.Index, normalizeBig(eta)) > 1e-12 {
		t.Error("dispatcher exact path disagrees with Exact")
	}
}

func sum(xs []float64) float64 {
	var s float64
	for _, x := range xs {
		s += x
	}
	return s
}
