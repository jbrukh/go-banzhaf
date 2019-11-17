package banzhaf

// Banzhaf returns the Banzhaf power index associated with a weighted voting
// system defined by the `weights` and `quota` provided. If `absolute` is set
// to true, then the absolute Banzhaf power index is returned.
func Banzhaf(weights []uint64, quota uint64, absolute bool) (index []float64, ok bool) {

	// get the total
	var total uint64
	for _, w := range weights {
		total += w
	}

	// check quota
	if quota > total || quota <= total/2 {
		return nil, false
	}

	// n
	n := uint64(len(weights))

	// polynomial
	polynomial := make([]uint64, total+1)
	polynomial[0] = 1

	// max order
	var order uint64

	// get polynomial weights
	for _, w := range weights {
		order += w
		offset := append(make([]uint64, w), polynomial...)
		for j := uint64(0); j <= order; j++ {
			polynomial[j] += offset[j]
		}
	}

	var (
		power   = make([]uint64, n)
		swings  = make([]uint64, quota)
		i, j, k uint64
	)

	for i = uint64(0); i < n; i++ {
		w := weights[i]
		for j = uint64(0); j < quota; j++ {
			if j < w {
				swings[j] = polynomial[j]
			} else {
				swings[j] = polynomial[j] - swings[j-w]
			}
		}
		for k = uint64(0); k < w; k++ {
			power[i] += swings[quota-1-k]
		}
	}

	var (
		denom uint64
		d     float64
	)

	// absolute Banzhaf power index
	if absolute {
		for _, c := range polynomial {
			denom += c
		}
		d = float64(denom) / 2
	} else { // normalized Banzhaf power index
		for _, p := range power {
			denom += p
		}
		d = float64(denom)
	}

	index = make([]float64, n)
	for i := range index {
		index[i] = float64(power[i]) / d
	}

	return index, true
}
