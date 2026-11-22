package decimal

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"testing"
)

// The reference model is math/big.Rat. For Add, Sub, Mul and Quo the result
// the package promises is the exact value rounded half to even to
//
//	scale = min(MaxScale, MaxPrec - n)
//
// fractional digits, where n is the number of digits of the integer part of
// the exact value (0 if |value| < 1). In words: 19 significant digits if the
// value has an integer part, otherwise 19 digits after the decimal point.
// If the rounded value is >= 10^19 (or n > 19), the result is ErrOverflow.
// Add, Sub, Mul round once, from the exact value. Quo has an extra step,
// see modelQuo.

var bigTen19 = new(big.Int).Exp(big.NewInt(10), big.NewInt(MaxPrec), nil)

func ratOf(d Decimal) *big.Rat {
	r, ok := new(big.Rat).SetString(d.String())
	if !ok {
		panic("bad decimal string " + d.String())
	}
	return r
}

// roundRatHalfEven rounds x as described above.
// It returns false if the integer part does not fit into MaxPrec digits.
func roundRatHalfEven(x *big.Rat) (*big.Rat, bool) {
	if x.Sign() == 0 {
		return new(big.Rat), true
	}
	a := new(big.Rat).Abs(x)
	ip := new(big.Int).Quo(a.Num(), a.Denom())
	n := 0
	if ip.Sign() > 0 {
		n = len(ip.String())
	}
	if n > MaxPrec {
		return nil, false
	}
	s := min(MaxScale, MaxPrec-n)
	pow := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(s)), nil)

	// q, r = divmod(a * 10^s)
	num := new(big.Int).Mul(a.Num(), pow)
	q, r := new(big.Int).QuoRem(num, a.Denom(), new(big.Int))
	switch new(big.Int).Lsh(r, 1).Cmp(a.Denom()) {
	case 1:
		q.Add(q, big.NewInt(1))
	case 0:
		if q.Bit(0) == 1 {
			q.Add(q, big.NewInt(1))
		}
	}
	res := new(big.Rat).SetFrac(q, pow)
	if res.Cmp(new(big.Rat).SetInt(bigTen19)) >= 0 {
		return nil, false
	}
	if x.Sign() < 0 {
		res.Neg(res)
	}
	return res, true
}

// modelQuo is the quotient the package computes. When the quotient is not
// exact, quoBint first truncates it to bscale (41) fractional digits and only
// then rounds half to even. That equals rounding the exact quotient once
// unless the exact quotient lies above a rounding midpoint h by less than
// 10^-41 (truncation would turn it into h itself). This cannot happen: for
// x = dcoef*10^(es-ds)/ecoef and a midpoint h = (2k+1)/(2*10^s),
// x-h is a nonzero multiple of 1/(2*10^s*ecoef*10^max(0, ds-es-s)), which is at
// least 1/(2*10^38), far above 10^-41, since ecoef < 10^19, ds <= 19 and s <= 19.
func modelQuo(x *big.Rat) (*big.Rat, bool) {
	return roundRatHalfEven(x)
}

// validDecimal checks the representation invariants of a result.
func validDecimal(t *testing.T, name string, got Decimal) {
	t.Helper()
	if got.neg && got.coef == 0 {
		t.Errorf("%v: negative zero %+v", name, got)
	}
	if got.coef > maxFint || got.scale < 0 || got.scale > MaxScale {
		t.Errorf("%v: invalid representation %+v", name, got)
	}
}

// checkResult compares the (got, err) pair of an operation with the model.
func checkResult(t *testing.T, name string, got Decimal, err error, exact *big.Rat, model func(*big.Rat) (*big.Rat, bool)) {
	t.Helper()
	want, ok := model(exact)
	if !ok {
		if err == nil {
			t.Errorf("%v = %q, want ErrOverflow (exact %v)", name, got, exact.FloatString(30))
		} else if !errors.Is(err, ErrOverflow) {
			t.Errorf("%v: error %q is not ErrOverflow", name, err)
		}
		return
	}
	if err != nil {
		t.Errorf("%v failed: %v (exact %v)", name, err, exact.FloatString(30))
		return
	}
	validDecimal(t, name, got)
	if ratOf(got).Cmp(want) != 0 {
		t.Errorf("%v = %q, want %v (exact %v)", name, got, want.FloatString(20), exact.FloatString(45))
	}
}

func checkAdd(t *testing.T, d, e Decimal) {
	t.Helper()
	got, err := d.Add(e)
	exact := new(big.Rat).Add(ratOf(d), ratOf(e))
	checkResult(t, "("+d.String()+").Add("+e.String()+")", got, err, exact, roundRatHalfEven)
}

func checkSub(t *testing.T, d, e Decimal) {
	t.Helper()
	got, err := d.Sub(e)
	exact := new(big.Rat).Sub(ratOf(d), ratOf(e))
	checkResult(t, "("+d.String()+").Sub("+e.String()+")", got, err, exact, roundRatHalfEven)
}

func checkMul(t *testing.T, d, e Decimal) {
	t.Helper()
	got, err := d.Mul(e)
	exact := new(big.Rat).Mul(ratOf(d), ratOf(e))
	checkResult(t, "("+d.String()+").Mul("+e.String()+")", got, err, exact, roundRatHalfEven)
}

func checkQuo(t *testing.T, d, e Decimal) {
	t.Helper()
	name := "(" + d.String() + ").Quo(" + e.String() + ")"
	got, err := d.Quo(e)
	if e.IsZero() {
		if !errors.Is(err, ErrDivisionByZero) {
			t.Errorf("%v: error %v, want ErrDivisionByZero", name, err)
		}
		return
	}
	exact := new(big.Rat).Quo(ratOf(d), ratOf(e))
	checkResult(t, name, got, err, exact, modelQuo)
}

func checkCmp(t *testing.T, d, e Decimal) {
	t.Helper()
	want := ratOf(d).Cmp(ratOf(e))
	if got := d.Cmp(e); got != want {
		t.Errorf("(%v).Cmp(%v) = %v, want %v", d, e, got, want)
	}
	if got := d.Equal(e); got != (want == 0) {
		t.Errorf("(%v).Equal(%v) = %v, want %v", d, e, got, want == 0)
	}
	if got := e.Cmp(d); got != -want {
		t.Errorf("(%v).Cmp(%v) = %v, want %v", e, d, got, -want)
	}
}

func newOrSkip(t *testing.T, coef int64, scale uint8) (Decimal, bool) {
	d, err := New(coef, int(scale))
	if err != nil {
		return Decimal{}, false
	}
	return d, true
}

// fuzzSeeds is the seed corpus: zeros, ones, int64 extremes (the largest
// coefficients New takes), scales 0 and 19, and mixed values.
var fuzzSeeds = []struct {
	coefD  int64
	scaleD uint8
	coefE  int64
	scaleE uint8
}{
	{0, 0, 0, 0},
	{1, 0, 1, 0},
	{-1, 0, 1, 0},
	{1, 0, -1, 0},
	{0, 19, 1, 19},
	{1, 19, 1, 19},
	{-1, 19, 1, 19},
	{math.MaxInt64, 0, 1, 0},
	{math.MaxInt64, 0, math.MaxInt64, 0},
	{math.MinInt64, 0, math.MinInt64, 0},
	{math.MinInt64, 0, 1, 0},
	{math.MaxInt64, 19, math.MaxInt64, 19},
	{math.MinInt64, 19, math.MaxInt64, 19},
	{math.MaxInt64, 19, 1, 0},
	{math.MaxInt64, 0, 1, 19},
	{1, 0, 3, 0},
	{2, 0, 3, 0},
	{1, 19, 3, 0},
	{5, 1, 3, 0},
	{15, 1, 1, 0},
	{12345, 4, 987654321, 9},
	{105, 0, 99999999999999999, 17},
	{math.MaxInt64, 1, math.MinInt64, 2},
}

func addSeeds(f *testing.F) {
	for _, s := range fuzzSeeds {
		f.Add(s.coefD, s.scaleD, s.coefE, s.scaleE)
	}
}

func FuzzAdd(f *testing.F) {
	addSeeds(f)
	f.Fuzz(func(t *testing.T, coefD int64, scaleD uint8, coefE int64, scaleE uint8) {
		d, ok := newOrSkip(t, coefD, scaleD)
		e, ok2 := newOrSkip(t, coefE, scaleE)
		if !ok || !ok2 {
			t.Skip()
		}
		checkAdd(t, d, e)
	})
}

func FuzzSub(f *testing.F) {
	addSeeds(f)
	f.Fuzz(func(t *testing.T, coefD int64, scaleD uint8, coefE int64, scaleE uint8) {
		d, ok := newOrSkip(t, coefD, scaleD)
		e, ok2 := newOrSkip(t, coefE, scaleE)
		if !ok || !ok2 {
			t.Skip()
		}
		checkSub(t, d, e)
	})
}

func FuzzMul(f *testing.F) {
	addSeeds(f)
	f.Fuzz(func(t *testing.T, coefD int64, scaleD uint8, coefE int64, scaleE uint8) {
		d, ok := newOrSkip(t, coefD, scaleD)
		e, ok2 := newOrSkip(t, coefE, scaleE)
		if !ok || !ok2 {
			t.Skip()
		}
		checkMul(t, d, e)
	})
}

func FuzzQuo(f *testing.F) {
	addSeeds(f)
	f.Fuzz(func(t *testing.T, coefD int64, scaleD uint8, coefE int64, scaleE uint8) {
		d, ok := newOrSkip(t, coefD, scaleD)
		e, ok2 := newOrSkip(t, coefE, scaleE)
		if !ok || !ok2 {
			t.Skip()
		}
		checkQuo(t, d, e)
	})
}

func FuzzCmp(f *testing.F) {
	addSeeds(f)
	f.Fuzz(func(t *testing.T, coefD int64, scaleD uint8, coefE int64, scaleE uint8) {
		d, ok := newOrSkip(t, coefD, scaleD)
		e, ok2 := newOrSkip(t, coefE, scaleE)
		if !ok || !ok2 {
			t.Skip()
		}
		checkCmp(t, d, e)
	})
}

// TestModelGrid runs the fuzz checks over a fixed grid that includes the
// coefficients up to maxFint, which int64 inputs of the fuzz targets cannot reach.
func TestModelGrid(t *testing.T) {
	grid := modelGrid()
	for _, d := range grid {
		for _, e := range grid {
			checkAdd(t, d, e)
			checkSub(t, d, e)
			checkMul(t, d, e)
			checkQuo(t, d, e)
			checkCmp(t, d, e)
		}
	}
}

// TestRoundModelGrid runs the rounding checks over the same grid.
func TestRoundModelGrid(t *testing.T) {
	grid := modelGrid()
	for _, d := range grid {
		for mode := Floor; mode <= HalfEven; mode++ {
			for target := -1; target <= 20; target++ {
				checkRound(t, d, target, mode)
			}
			for _, step := range grid {
				checkRoundStep(t, d, step, mode)
			}
		}
		for _, step := range grid {
			checkIsMultipleOf(t, d, step)
		}
	}
}

func modelGrid() []Decimal {
	var grid []Decimal
	for _, s := range []string{
		"0", "0.0000000000000000000", "1", "-1", "3", "7", "0.5", "1.5", "2.5", "-2.5",
		"9999999999999999999", "-9999999999999999999",
		"9999999999999999998", "1000000000000000000", "5000000000000000000",
		"0.9999999999999999999", "-0.9999999999999999999", "0.0000000000000000001",
		"-0.0000000000000000001", "0.5000000000000000000", "0.1234567890123456789",
		"1234567890123456789", "1.234567890123456789", "12345678901234567.89",
		"9.999999999999999999", "99999999999999999.99", "0.3333333333333333333",
		"1.000000000000000001", "100", "0.01", "3.00",
	} {
		grid = append(grid, MustParse(s))
	}
	return grid
}

// roundRatInt rounds x to an integer using the mode.
func roundRatInt(x *big.Rat, mode RoundingMode) *big.Int {
	fl := new(big.Int).Div(x.Num(), x.Denom()) // Euclidean, Denom > 0: floor
	if x.IsInt() {
		return fl
	}
	ceil := new(big.Int).Add(fl, big.NewInt(1))
	switch mode {
	case Floor:
		return fl
	case Ceil:
		return ceil
	case Down:
		if x.Sign() >= 0 {
			return fl
		}
		return ceil
	case Up:
		if x.Sign() >= 0 {
			return ceil
		}
		return fl
	}
	frac := new(big.Rat).Sub(x, new(big.Rat).SetInt(fl))
	switch c := frac.Cmp(big.NewRat(1, 2)); {
	case c < 0:
		return fl
	case c > 0:
		return ceil
	}
	// exactly half
	if mode == HalfUp {
		if x.Sign() >= 0 {
			return ceil
		}
		return fl
	}
	if fl.Bit(0) == 0 {
		return fl
	}
	return ceil
}

func ratPow10(n int) *big.Rat {
	return new(big.Rat).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(n)), nil))
}

func checkRound(t *testing.T, d Decimal, target int, mode RoundingMode) {
	t.Helper()
	name := fmt.Sprintf("(%v).Round(%v, %v)", d, target, mode)
	got := d.Round(target, mode)
	validDecimal(t, name, got)
	scale := min(max(target, 0), MaxScale)
	if d.Scale() <= scale {
		if got != d {
			t.Errorf("%v = %+v, want d unchanged", name, got)
		}
		return
	}
	p := ratPow10(scale)
	n := roundRatInt(new(big.Rat).Mul(ratOf(d), p), mode)
	want := new(big.Rat).Quo(new(big.Rat).SetInt(n), p)
	if ratOf(got).Cmp(want) != 0 || got.Scale() != scale {
		t.Errorf("%v = %q (scale %v), want %v at scale %v", name, got, got.Scale(), want.FloatString(20), scale)
	}
}

func checkRoundStep(t *testing.T, d, step Decimal, mode RoundingMode) {
	t.Helper()
	name := fmt.Sprintf("(%v).RoundStep(%v, %v)", d, step, mode)
	got, err := d.RoundStep(step, mode)
	if step.Sign() <= 0 {
		if !errors.Is(err, ErrInvalidStep) {
			t.Errorf("%v: error %v, want ErrInvalidStep", name, err)
		}
		return
	}
	n := roundRatInt(new(big.Rat).Quo(ratOf(d), ratOf(step)), mode)
	coef := new(big.Int).Mul(n, new(big.Int).SetUint64(uint64(step.coef)))
	if new(big.Int).Abs(coef).Cmp(new(big.Int).SetUint64(maxFint)) > 0 {
		if !errors.Is(err, ErrOverflow) {
			t.Errorf("%v = %v, %v, want ErrOverflow", name, got, err)
		}
		return
	}
	if err != nil {
		t.Errorf("%v failed: %v", name, err)
		return
	}
	validDecimal(t, name, got)
	want := new(big.Rat).Mul(new(big.Rat).SetInt(n), ratOf(step))
	if ratOf(got).Cmp(want) != 0 || got.Scale() != step.Scale() {
		t.Errorf("%v = %q (scale %v), want %v at scale %v", name, got, got.Scale(), want.FloatString(20), step.Scale())
	}
}

func checkIsMultipleOf(t *testing.T, d, step Decimal) {
	t.Helper()
	want := false
	if step.Sign() > 0 {
		want = new(big.Rat).Quo(ratOf(d), ratOf(step)).IsInt()
	}
	if got := d.IsMultipleOf(step); got != want {
		t.Errorf("(%v).IsMultipleOf(%v) = %v, want %v", d, step, got, want)
	}
}

func FuzzRound(f *testing.F) {
	for _, s := range fuzzSeeds {
		for mode := Floor; mode <= HalfEven; mode++ {
			f.Add(s.coefD, s.scaleD, s.scaleE, uint8(mode))
		}
	}
	f.Fuzz(func(t *testing.T, coef int64, scale, target, mode uint8) {
		d, ok := newOrSkip(t, coef, scale)
		if !ok {
			t.Skip()
		}
		checkRound(t, d, int(target), RoundingMode(mode%6))
	})
}

func FuzzRoundStep(f *testing.F) {
	for _, s := range fuzzSeeds {
		for mode := Floor; mode <= HalfEven; mode++ {
			f.Add(s.coefD, s.scaleD, s.coefE, s.scaleE, uint8(mode))
		}
	}
	f.Fuzz(func(t *testing.T, coefD int64, scaleD uint8, coefS int64, scaleS, mode uint8) {
		d, ok := newOrSkip(t, coefD, scaleD)
		step, ok2 := newOrSkip(t, coefS, scaleS)
		if !ok || !ok2 {
			t.Skip()
		}
		checkRoundStep(t, d, step, RoundingMode(mode%6))
	})
}

func FuzzIsMultipleOf(f *testing.F) {
	addSeeds(f)
	f.Fuzz(func(t *testing.T, coefD int64, scaleD uint8, coefS int64, scaleS uint8) {
		d, ok := newOrSkip(t, coefD, scaleD)
		step, ok2 := newOrSkip(t, coefS, scaleS)
		if !ok || !ok2 {
			t.Skip()
		}
		checkIsMultipleOf(t, d, step)
	})
}
