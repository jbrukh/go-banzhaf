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

}

func testBanzhaf(t *testing.T, weights []uint64, quota uint64, absolute bool, want []float64) {
	got, ok := Banzhaf(weights, quota, absolute)

	if !ok {
		t.Errorf("function returned an error")
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got=%v, want=%v, weights=%v, quota=%v, absolute=%v", got, want, weights, quota, absolute)
	}
}
