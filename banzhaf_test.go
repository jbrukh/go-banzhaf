package banzhaf

import (
	"reflect"
	"testing"
)

func TestBasic(t *testing.T) {
	t.Run("basic absolute", func(t *testing.T) {
		var (
			weights  = []uint64{2, 2, 1}
			quota    = uint64(4)
			absolute = true
			want     = []float64{0.5, 0.5, 0}
		)
		testBanzhaf(t, weights, quota, absolute, want)
	})

	t.Run("basic normalized", func(t *testing.T) {
		var (
			weights  = []uint64{2, 2, 1}
			quota    = uint64(4)
			absolute = false
			want     = []float64{0.5, 0.5, 0}
		)
		testBanzhaf(t, weights, quota, absolute, want)
	})

	t.Run("case 2", func(t *testing.T) {
		var (
			weights  = []uint64{2, 2, 2, 1}
			quota    = uint64(4)
			absolute = true
			want     = []float64{0.5, 0.5, 0.5, 0}
		)
		testBanzhaf(t, weights, quota, absolute, want)
	})

	t.Run("case 3", func(t *testing.T) {
		var (
			weights  = []uint64{3, 2, 2, 1}
			quota    = uint64(5)
			absolute = true
			want     = []float64{0.625, 0.375, 0.375, 0.125}
		)
		testBanzhaf(t, weights, quota, absolute, want)
	})

	t.Run("case 4", func(t *testing.T) {
		var (
			weights  = []uint64{3, 2, 2, 1}
			quota    = uint64(5)
			absolute = true
			want     = []float64{0.625, 0.375, 0.375, 0.125}
		)
		testBanzhaf(t, weights, quota, absolute, want)
	})

	t.Run("quota less than half", func(t *testing.T) {
		var (
			weights  = []uint64{3, 2, 2, 1}
			quota    = uint64(3)
			absolute = true
		)
		testBanzhafErr(t, weights, quota, absolute)
	})

	t.Run("quota equal half", func(t *testing.T) {
		var (
			weights  = []uint64{3, 2, 2, 1}
			quota    = uint64(4)
			absolute = true
		)
		testBanzhafErr(t, weights, quota, absolute)
	})

	t.Run("quota equal total", func(t *testing.T) {
		var (
			weights  = []uint64{3, 2, 2, 1}
			quota    = uint64(8)
			absolute = true
			want     = []float64{0.125, 0.125, 0.125, 0.125}
		)
		testBanzhaf(t, weights, quota, absolute, want)
	})

	t.Run("quota long array", func(t *testing.T) {
		var (
			n        = 10000
			quota    = uint64(n/2 + 1)
			absolute = false
			weights  []uint64
			want     []float64
		)
		for i := 0; i < n; i++ {
			weights = append(weights, 1)
			want = append(want, 0.0001)
		}
		testBanzhaf(t, weights, quota, absolute, want)
	})

}

func testBanzhaf(t *testing.T, weights []uint64, quota uint64, absolute bool, want []float64) {
	got, err := Banzhaf(weights, quota, absolute)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got=%v, want=%v, weights=%v, quota=%v, absolute=%v", got, want, weights, quota, absolute)
	}
}

func testBanzhafErr(t *testing.T, weights []uint64, quota uint64, absolute bool) {
	_, err := Banzhaf(weights, quota, absolute)
	if err == nil {
		t.Errorf("expecting an error but got: %v", err)
	}
}
