package banzhaf

import (
	"fmt"
	"log"
	"math/big"
)

// Banzhaf returns the Banzhaf power index associated with a weighted voting
// system defined by the `weights` and `quota` provided. If `absolute` is set
// to true, then the absolute Banzhaf power index is returned.
func Banzhaf(weights []uint64, quota uint64, absolute bool) (index []float64, err error) {

	var (
		total      uint64   // total votes
		n          uint64   // number of players
		order      uint64   // maximum order of the polynomial
		polynomial []uint64 // polynomial generator
		i, j, k    uint64   // indices
	)

	// calculate the total votes
	for _, w := range weights {
		total += w
	}

	// check quota
	if quota > total || quota <= total/2 {
		return nil, fmt.Errorf("the quota is out of bounds: (%d,%d]", total, total/2)
	}

	// n
	n = uint64(len(weights))

	// polynomial
	polynomial = make([]uint64, total+1)
	polynomial[0] = 1

	// get polynomial weights
	for _, w := range weights {
		order += w
		offset := append(make([]uint64, w), polynomial...)
		for j = 0; j <= order; j++ {
			polynomial[j] += offset[j]
		}
	}

	log.Printf("poly=%v\n", polynomial)

	var (
		// an array counting Banzhaf power (swings)
		power = make([]uint64, n)

		// an array counting all swings
		swings = make([]uint64, quota)

		// denominator for the power index
		denom = big.NewInt(0)
	)

	// count swings and banzhaf power
	for i = 0; i < n; i++ {
		w := weights[i]
		for j = 0; j < quota; j++ {
			if j < w {
				swings[j] = polynomial[j]
			} else {
				swings[j] = polynomial[j] - swings[j-w]
			}
		}
		for k = 0; k < w; k++ {
			power[i] += swings[quota-1-k]
		}
	}

	if absolute {
		// absolute Banzhaf power index takes the
		// denominator as all possible votes where
		// everyone else other than this player participates
		// which is 2^(n-1)
		denom.Exp(big.NewInt(2), new(big.Int).SetUint64(n-1), nil)
		log.Printf("l=%d, d=%v\n", len(polynomial), denom)
	} else {
		// normalized Banzhaf power index takes the
		// denominator as all possible swings
		for _, p := range power {
			denom.Add(denom, new(big.Int).SetUint64(p))
		}
	}

	index = make([]float64, n)
	d := new(big.Float).SetInt(denom)
	for i := range index {
		p := new(big.Float).SetUint64(power[i])
		index[i], _ = new(big.Float).Quo(p, d).Float64()
	}

	return index, nil
}
