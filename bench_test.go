package banzhaf

import (
	"math"
	"math/rand"
	"testing"
)

// zipfWeights builds a power-law holder snapshot scaled to raw on-chain units.
func zipfWeights(n int, alpha float64, scale, decimals float64, seed int64) []float64 {
	rng := rand.New(rand.NewSource(seed))
	w := make([]float64, n)
	for i := 0; i < n; i++ {
		v := scale / math.Pow(float64(i+1), alpha)
		if v < 1 {
			v = 1
		}
		w[i] = v * decimals
	}
	rng.Shuffle(n, func(i, j int) { w[i], w[j] = w[j], w[i] })
	return w
}

func BenchmarkHighN100k(b *testing.B) {
	w := zipfWeights(100_000, 1.0, 1e6, 1e18, 0)
	var total float64
	for _, x := range w {
		total += x
	}
	q := total/2 + 1
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		HighN(w, q, HighNOptions{})
	}
}

func BenchmarkHighN250k(b *testing.B) {
	w := zipfWeights(250_000, 1.0, 1e6, 1e18, 0)
	var total float64
	for _, x := range w {
		total += x
	}
	q := total/2 + 1
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		HighN(w, q, HighNOptions{})
	}
}
