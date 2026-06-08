package banzhaf

import (
	"encoding/json"
	"math"
	"os"
	"sort"
	"testing"
)

// vectorCase mirrors one entry of testdata/vectors.json, produced by the Python
// reference implementation. The Go results must reproduce these.
type vectorCase struct {
	Name      string    `json:"name"`
	Weights   []float64 `json:"weights"`
	Quota     uint64    `json:"quota"`
	ExactNorm []float64 `json:"exact_norm"`
	ExactAbs  []float64 `json:"exact_abs"`
	HighNNorm []float64 `json:"highn_norm"`
}

func loadVectors(t *testing.T) []vectorCase {
	t.Helper()
	b, err := os.ReadFile("testdata/vectors.json")
	if err != nil {
		t.Fatalf("read vectors: %v", err)
	}
	var v struct {
		Cases []vectorCase `json:"cases"`
	}
	if err := json.Unmarshal(b, &v); err != nil {
		t.Fatalf("parse vectors: %v", err)
	}
	if len(v.Cases) == 0 {
		t.Fatal("no cases in vectors.json")
	}
	return v.Cases
}

func maxAbsDiff(a, b []float64) float64 {
	var m float64
	for i := range a {
		if d := math.Abs(a[i] - b[i]); d > m {
			m = d
		}
	}
	return m
}

func sortedDesc(a []float64) []float64 {
	c := append([]float64(nil), a...)
	sort.Sort(sort.Reverse(sort.Float64Slice(c)))
	return c
}

// TestExactMatchesPython proves the Go exact grouped method reproduces the Python
// reference bit-for-bit (up to float rounding) on both the normalized and the
// absolute index.
func TestExactMatchesPython(t *testing.T) {
	for _, c := range loadVectors(t) {
		uw := make([]uint64, len(c.Weights))
		for i, w := range c.Weights {
			uw[i] = uint64(w)
		}
		eta, err := Exact(uw, c.Quota)
		if err != nil {
			t.Fatalf("%s: Exact: %v", c.Name, err)
		}
		if d := maxAbsDiff(normalizeBig(eta), c.ExactNorm); d > 1e-12 {
			t.Errorf("%s: normalized index differs from Python by %.2e", c.Name, d)
		}
		if len(c.Weights) <= 30 { // avoid 2^(n-1) underflow noise for big n
			if d := maxAbsDiff(absoluteBig(eta), c.ExactAbs); d > 1e-12 {
				t.Errorf("%s: absolute index differs from Python by %.2e", c.Name, d)
			}
		}
	}
}

// TestHighNMatchesPython proves the Go high-n approximation reproduces the Python
// one. Equal-weight holders may swap across the head/tail boundary between the two
// implementations, so we compare the sorted index distribution.
func TestHighNMatchesPython(t *testing.T) {
	for _, c := range loadVectors(t) {
		got := Normalize(HighN(c.Weights, float64(c.Quota), HighNOptions{}))
		d := maxAbsDiff(sortedDesc(got), sortedDesc(c.HighNNorm))
		if d > 1e-9 {
			t.Errorf("%s: high-n index differs from Python by %.2e", c.Name, d)
		}
	}
}

// TestHighNAccuracyVsExact confirms the approximation is close to the exact answer
// on the reference cases (the property that actually matters to a user).
func TestHighNAccuracyVsExact(t *testing.T) {
	for _, c := range loadVectors(t) {
		got := Normalize(HighN(c.Weights, float64(c.Quota), HighNOptions{}))
		if d := maxAbsDiff(sortedDesc(got), sortedDesc(c.ExactNorm)); d > 5e-3 {
			t.Errorf("%s: high-n vs exact max|dbeta| = %.2e", c.Name, d)
		}
	}
}
