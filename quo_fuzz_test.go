package decimal

import (
	"errors"
	"fmt"
	"math/big"
	"testing"
)

// checkQuoRound compares QuoRound with the exact big.Rat quotient:
// n is the number of digits of its integer part, the scale is
// s = min(scale, MaxPrec-n) and the value is rounded to s digits by the mode.
// If that rounding carries to 10^19 the result is 10^(19-s) with scale s-1,
// or ErrOverflow when s is 0.
func checkQuoRound(t *testing.T, d, e Decimal, scale int, mode RoundingMode) {
	t.Helper()
	name := fmt.Sprintf("(%v).QuoRound(%v, %v, %v)", d, e, scale, mode)
	got, err := d.QuoRound(e, scale, mode)
	if e.IsZero() {
		if !errors.Is(err, ErrDivisionByZero) {
			t.Errorf("%v: error %v, want ErrDivisionByZero", name, err)
		}
		return
	}
	x := new(big.Rat).Quo(ratOf(d), ratOf(e))
	ip := new(big.Int).Quo(new(big.Rat).Abs(x).Num(), new(big.Rat).Abs(x).Denom())
	n := 0
	if ip.Sign() > 0 {
		n = len(ip.String())
	}
	if n > MaxPrec {
		if !errors.Is(err, ErrOverflow) {
			t.Errorf("%v = %v, %v, want ErrOverflow", name, got, err)
		}
		return
	}
	s := min(max(scale, 0), MaxScale, MaxPrec-n)
	N := roundRatInt(new(big.Rat).Mul(x, ratPow10(s)), mode)
	if new(big.Int).Abs(N).Cmp(bigTen19) >= 0 {
		if s == 0 {
			if !errors.Is(err, ErrOverflow) {
				t.Errorf("%v = %v, %v, want ErrOverflow (carry)", name, got, err)
			}
			return
		}
		N.Quo(N, big.NewInt(10))
		s--
	}
	if err != nil {
		t.Errorf("%v failed: %v (x = %v)", name, err, x.FloatString(30))
		return
	}
	validDecimal(t, name, got)
	want := new(big.Rat).Quo(new(big.Rat).SetInt(N), ratPow10(s))
	if ratOf(got).Cmp(want) != 0 || got.Scale() != s {
		t.Errorf("%v = %q (scale %v), want %v at scale %v (x = %v)", name, got, got.Scale(), want.FloatString(20), s, x.FloatString(30))
	}
}

func FuzzQuoRound(f *testing.F) {
	for i, s := range fuzzSeeds {
		for mode := Floor; mode <= HalfEven; mode++ {
			f.Add(s.coefD, s.scaleD, s.coefE, s.scaleE, uint8([]int{0, 2, 16, 19, 8}[i%5]), uint8(mode))
		}
	}
	f.Fuzz(func(t *testing.T, coefD int64, scaleD uint8, coefE int64, scaleE, scale, mode uint8) {
		d, ok := newOrSkip(t, coefD, scaleD)
		e, ok2 := newOrSkip(t, coefE, scaleE)
		if !ok || !ok2 {
			t.Skip()
		}
		checkQuoRound(t, d, e, int(scale), RoundingMode(mode%6))
	})
}

// TestQuoRoundModelGrid runs the check over the fixed grid of 19-digit values.
func TestQuoRoundModelGrid(t *testing.T) {
	grid := modelGrid()
	for _, d := range grid {
		for _, e := range grid {
			for mode := Floor; mode <= HalfEven; mode++ {
				for _, scale := range []int{-1, 0, 1, 2, 5, 16, 18, 19, 20} {
					checkQuoRound(t, d, e, scale, mode)
				}
			}
		}
	}
}
