// Derived from github.com/govalues/decimal v0.1.36 (MIT), see LICENSE.
package decimal

import (
	"math"
	"testing"
)

func TestFint_add(t *testing.T) {
	cases := []struct {
		x, y, wantCoef fint
		wantOk         bool
	}{
		{0, 0, 0, true},
		{maxFint, 1, 0, false},
	}
	for _, tt := range cases {
		x, y := tt.x, tt.y
		gotCoef, gotOk := x.add(y)
		if gotCoef != tt.wantCoef || gotOk != tt.wantOk {
			t.Errorf("%v.add(%v) = %v, %v, want %v, %v", x, y, gotCoef, gotOk, tt.wantCoef, tt.wantOk)
		}
	}
}

func TestFint_mul(t *testing.T) {
	cases := []struct {
		x, y, wantCoef fint
		wantOk         bool
	}{
		{0, 0, 0, true},
		{10, 10, 100, true},
		{maxFint, 2, 0, false},
	}
	for _, tt := range cases {
		x, y := tt.x, tt.y
		gotCoef, gotOk := x.mul(y)
		if gotCoef != tt.wantCoef || gotOk != tt.wantOk {
			t.Errorf("%v.mul(%v) = %v, %v, want %v, %v", x, y, gotCoef, gotOk, tt.wantCoef, tt.wantOk)
		}
	}
}

func TestFint_quo(t *testing.T) {
	cases := []struct {
		x, y, wantCoef fint
		wantOk         bool
	}{
		{1, 0, 0, false},
		{1, 1, 1, true},
		{2, 4, 0, false},
		{20, 4, 5, true},
		{20, 3, 0, false},
		{maxFint, 2, 0, false},
	}
	for _, tt := range cases {
		x, y := tt.x, tt.y
		gotCoef, gotOk := x.quo(y)
		if gotCoef != tt.wantCoef || gotOk != tt.wantOk {
			t.Errorf("%v.quo(%v) = %v, %v, want %v, %v", x, y, gotCoef, gotOk, tt.wantCoef, tt.wantOk)
		}
	}
}

func TestFint_quoRem(t *testing.T) {
	cases := []struct {
		x, y, wantQuo, wantRem fint
		wantOk                 bool
	}{
		{1, 0, 0, 0, false},
		{1, 1, 1, 0, true},
		{2, 4, 0, 2, true},
		{20, 4, 5, 0, true},
		{20, 3, 6, 2, true},
		{maxFint, 2, maxFint / 2, 1, true},
	}
	for _, tt := range cases {
		x, y := tt.x, tt.y
		gotQuo, gotRem, gotOk := x.quoRem(y)
		if gotQuo != tt.wantQuo || gotRem != tt.wantRem || gotOk != tt.wantOk {
			t.Errorf("%v.quoRem(%v) = %v, %v, %v, want %v, %v, %v", x, y, gotQuo, gotRem, gotOk, tt.wantQuo, tt.wantRem, tt.wantOk)
		}
	}
}

func TestFint_subAbs(t *testing.T) {
	cases := []struct {
		x, y, wantCoef fint
	}{
		{1, 0, 1},
		{1, 1, 0},
		{2, 4, 2},
		{20, 4, 16},
		{20, 3, 17},
		{maxFint, 2, 9_999_999_999_999_999_997},
		{maxFint, 1, 9_999_999_999_999_999_998},
		{maxFint, 0, maxFint},
	}
	for _, tt := range cases {
		x, y := tt.x, tt.y
		gotCoef := x.subAbs(y)
		if gotCoef != tt.wantCoef {
			t.Errorf("%v.subAbs(%v) = %v, want %v", x, y, gotCoef, tt.wantCoef)
		}
	}
}

func TestFint_lsh(t *testing.T) {
	cases := []struct {
		x        fint
		shift    int
		wantCoef fint
		wantOk   bool
	}{
		{0, 0, 0, true},
		{10, 0, 10, true},
		{0, 1, 0, true},
		{10, 1, 100, true},
		{1, 20, 0, false},
	}
	for _, tt := range cases {
		x := tt.x
		gotCoef, gotOk := x.lsh(tt.shift)
		if gotCoef != tt.wantCoef || gotOk != tt.wantOk {
			t.Errorf("%v.lsh(%v) = %v, %v, want %v, %v", x, tt.shift, gotCoef, gotOk, tt.wantCoef, tt.wantOk)
		}
	}
}

func TestFint_fsa(t *testing.T) {
	cases := []struct {
		x        fint
		shift    int
		y        byte
		wantCoef fint
		wantOk   bool
	}{
		{0, 0, 0, 0, true},
		{1, 1, 1, 11, true},
		{2, 2, 2, 202, true},
		{3, 3, 3, 3_003, true},
		{1, 20, 0, 0, false},
		{maxFint, 0, 1, 0, false},
	}
	for _, tt := range cases {
		x, y, shift := tt.x, tt.y, tt.shift
		gotCoef, gotOk := x.fsa(shift, y)
		if gotCoef != tt.wantCoef || gotOk != tt.wantOk {
			t.Errorf("%v.fsa(%v, %v) = %v, %v, want %v, %v", x, shift, y, gotCoef, gotOk, tt.wantCoef, tt.wantOk)
		}
	}
}

func TestFint_rshHalfEven(t *testing.T) {
	cases := []struct {
		x     fint
		shift int
		want  fint
	}{
		// Negative shift
		{1, -1, 1},

		// Rounding
		{1, 0, 1},
		{20, 1, 2},
		{18, 1, 2},
		{15, 1, 2},
		{12, 1, 1},
		{10, 1, 1},
		{8, 1, 1},
		{5, 1, 0},
		{2, 1, 0},
		{maxFint, 19, 1},

		// Large shifts
		{0, 17, 0},
		{0, 18, 0},
		{0, 19, 0},
		{0, 20, 0},
		{0, 21, 0},

		{1, 17, 0},
		{1, 18, 0},
		{1, 19, 0},
		{1, 20, 0},
		{1, 21, 0},

		{5_000_000_000_000_000_000, 17, 50},
		{5_000_000_000_000_000_000, 18, 5},
		{5_000_000_000_000_000_000, 19, 0},
		{5_000_000_000_000_000_000, 20, 0},
		{5_000_000_000_000_000_000, 21, 0},

		{5_000_000_000_000_000_001, 17, 50},
		{5_000_000_000_000_000_001, 18, 5},
		{5_000_000_000_000_000_001, 19, 1},
		{5_000_000_000_000_000_001, 20, 0},
		{5_000_000_000_000_000_001, 21, 0},

		{maxFint, 17, 100},
		{maxFint, 18, 10},
		{maxFint, 19, 1},
		{maxFint, 20, 0},
		{maxFint, 21, 0},

		{10_000_000_000_000_000_000, 17, 100},
		{10_000_000_000_000_000_000, 18, 10},
		{10_000_000_000_000_000_000, 19, 1},
		{10_000_000_000_000_000_000, 20, 0},
		{10_000_000_000_000_000_000, 21, 0},

		{14_999_999_999_999_999_999, 17, 150},
		{14_999_999_999_999_999_999, 18, 15},
		{14_999_999_999_999_999_999, 19, 1},
		{14_999_999_999_999_999_999, 20, 0},
		{14_999_999_999_999_999_999, 21, 0},

		{15_000_000_000_000_000_000, 17, 150},
		{15_000_000_000_000_000_000, 18, 15},
		{15_000_000_000_000_000_000, 19, 2},
		{15_000_000_000_000_000_000, 20, 0},
		{15_000_000_000_000_000_000, 21, 0},

		{math.MaxUint64, 17, 184},
		{math.MaxUint64, 18, 18},
		{math.MaxUint64, 19, 2},
		{math.MaxUint64, 20, 0},
		{math.MaxUint64, 21, 0},
	}
	for _, tt := range cases {
		got := tt.x.rshHalfEven(tt.shift)
		if got != tt.want {
			t.Errorf("%v.rshHalfEven(%v) = %v, want %v", tt.x, tt.shift, got, tt.want)
		}
	}
}

func TestFint_rshDown(t *testing.T) {
	cases := []struct {
		x     fint
		shift int
		want  fint
	}{
		// Negative shift
		{1, -1, 1},
		// Rounding
		{1, 0, 1},
		{20, 1, 2},
		{18, 1, 1},
		{15, 1, 1},
		{12, 1, 1},
		{10, 1, 1},
		{8, 1, 0},
		{5, 1, 0},
		{2, 1, 0},

		// Large shifts
		{0, 17, 0},
		{0, 18, 0},
		{0, 19, 0},
		{0, 20, 0},
		{0, 21, 0},

		{1, 17, 0},
		{1, 18, 0},
		{1, 19, 0},
		{1, 20, 0},
		{1, 21, 0},

		{maxFint, 17, 99},
		{maxFint, 18, 9},
		{maxFint, 19, 0},
		{maxFint, 20, 0},
		{maxFint, 21, 0},

		{10_000_000_000_000_000_000, 17, 100},
		{10_000_000_000_000_000_000, 18, 10},
		{10_000_000_000_000_000_000, 19, 1},
		{10_000_000_000_000_000_000, 20, 0},
		{10_000_000_000_000_000_000, 21, 0},

		{math.MaxUint64, 17, 184},
		{math.MaxUint64, 18, 18},
		{math.MaxUint64, 19, 1},
		{math.MaxUint64, 20, 0},
		{math.MaxUint64, 21, 0},
	}
	for _, tt := range cases {
		got := tt.x.rshDown(tt.shift)
		if got != tt.want {
			t.Errorf("%v.rshDown(%v) = %v, want %v", tt.x, tt.shift, got, tt.want)
		}
	}
}

func TestFint_prec(t *testing.T) {
	cases := []struct {
		x    fint
		want int
	}{
		{0, 0},
		{1, 1},
		{9, 1},
		{10, 2},
		{99, 2},
		{100, 3},
		{999, 3},
		{1_000, 4},
		{9_999, 4},
		{10_000, 5},
		{99_999, 5},
		{100_000, 6},
		{999_999, 6},
		{1_000_000, 7},
		{9_999_999, 7},
		{10_000_000, 8},
		{99_999_999, 8},
		{100_000_000, 9},
		{999_999_999, 9},
		{1_000_000_000, 10},
		{9_999_999_999, 10},
		{10_000_000_000, 11},
		{99_999_999_999, 11},
		{100_000_000_000, 12},
		{999_999_999_999, 12},
		{1_000_000_000_000, 13},
		{9_999_999_999_999, 13},
		{10_000_000_000_000, 14},
		{99_999_999_999_999, 14},
		{100_000_000_000_000, 15},
		{999_999_999_999_999, 15},
		{1_000_000_000_000_000, 16},
		{9_999_999_999_999_999, 16},
		{10_000_000_000_000_000, 17},
		{99_999_999_999_999_999, 17},
		{100_000_000_000_000_000, 18},
		{999_999_999_999_999_999, 18},
		{1_000_000_000_000_000_000, 19},
		{maxFint, 19},
		{10_000_000_000_000_000_000, 20},
		{math.MaxUint64, 20},
	}
	for _, tt := range cases {
		got := tt.x.prec()
		if got != tt.want {
			t.Errorf("%v.prec() = %v, want %v", tt.x, got, tt.want)
		}
	}
}

func TestFint_ntz(t *testing.T) {
	cases := []struct {
		x    fint
		want int
	}{
		{0, 0},
		{1, 0},
		{9, 0},
		{10, 1},
		{99, 0},
		{100, 2},
		{999, 0},
		{1_000, 3},
		{9_999, 0},
		{10_000, 4},
		{99_999, 0},
		{100_000, 5},
		{999_999, 0},
		{1_000_000, 6},
		{9_999_999, 0},
		{10_000_000, 7},
		{99_999_999, 0},
		{100_000_000, 8},
		{999_999_999, 0},
		{1_000_000_000, 9},
		{9_999_999_999, 0},
		{10_000_000_000, 10},
		{99_999_999_999, 0},
		{100_000_000_000, 11},
		{999_999_999_999, 0},
		{1_000_000_000_000, 12},
		{9_999_999_999_999, 0},
		{10_000_000_000_000, 13},
		{99_999_999_999_999, 0},
		{100_000_000_000_000, 14},
		{999_999_999_999_999, 0},
		{1_000_000_000_000_000, 15},
		{9_999_999_999_999_999, 0},
		{10_000_000_000_000_000, 16},
		{99_999_999_999_999_999, 0},
		{100_000_000_000_000_000, 17},
		{999_999_999_999_999_999, 0},
		{1_000_000_000_000_000_000, 18},
		{1_000_000_000_000_000_001, 0},
		{maxFint, 0},
		{10_000_000_000_000_000_000, 19},
		{math.MaxUint64, 0},
	}
	for _, tt := range cases {
		got := tt.x.ntz()
		if got != tt.want {
			t.Errorf("%v.ntz() = %v, want %v", tt.x, got, tt.want)
		}
	}
}

func TestFint_hasPrec(t *testing.T) {
	cases := []struct {
		x    fint
		prec int
		want bool
	}{
		{0, 0, true},
		{0, 1, false},
		{0, 2, false},

		{1, 0, true},
		{1, 1, true},
		{1, 2, false},

		{9, 0, true},
		{9, 1, true},
		{9, 2, false},

		{10, 0, true},
		{10, 1, true},
		{10, 2, true},
		{10, 3, false},

		{99, 0, true},
		{99, 1, true},
		{99, 2, true},
		{99, 3, false},

		{100_000_000_000_000, 17, false},
		{1_000_000_000_000_000, 17, false},
		{10_000_000_000_000_000, 17, true},
		{100_000_000_000_000_000, 17, true},
		{1_000_000_000_000_000_000, 17, true},

		{1_000_000_000_000_000_000, 17, true},
		{1_000_000_000_000_000_000, 18, true},
		{1_000_000_000_000_000_000, 19, true},
		{1_000_000_000_000_000_000, 20, false},
		{1_000_000_000_000_000_000, 21, false},

		{maxFint, 17, true},
		{maxFint, 18, true},
		{maxFint, 19, true},
		{maxFint, 20, false},
		{maxFint, 21, false},

		{10_000_000_000_000_000_000, 17, true},
		{10_000_000_000_000_000_000, 18, true},
		{10_000_000_000_000_000_000, 19, true},
		{10_000_000_000_000_000_000, 20, true},
		{10_000_000_000_000_000_000, 21, false},

		{math.MaxUint64, 17, true},
		{math.MaxUint64, 18, true},
		{math.MaxUint64, 19, true},
		{math.MaxUint64, 20, true},
		{math.MaxUint64, 21, false},
	}
	for _, tt := range cases {
		got := tt.x.hasPrec(tt.prec)
		if got != tt.want {
			t.Errorf("%v.hasPrec(%v) = %v, want %v", tt.x, tt.prec, got, tt.want)
		}
	}
}
