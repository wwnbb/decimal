package decimal

import (
	"fmt"
	"math/bits"
)

// QuoRound returns the quotient d / e rounded to scale digits after the
// decimal point using the mode. The scale is clamped to [0, MaxScale].
//
// The rounding is decided from the exact remainder of the division, never
// from a quotient already rounded to 19 digits, so it is exact for every
// operand. If the integer part of the quotient is so long that fewer than
// scale fractional digits fit into [MaxPrec] significant digits, the scale is
// reduced to what fits (12345678901234 / 3 with scale 16 has scale 6) and the
// rounding uses the same mode at that scale. The same happens when rounding
// carries into a new digit (9.9..9 up to 10), which keeps the value exact.
// A result that rounds to zero is zero, never negative.
//
// QuoRound returns an error matching [ErrDivisionByZero] if e is zero,
// [ErrInvalidMode] if the mode is unknown, and [ErrOverflow] if the integer
// part of the result needs more than [MaxPrec] digits.
func (d Decimal) QuoRound(e Decimal, scale int, mode RoundingMode) (Decimal, error) {
	if e.IsZero() {
		return Decimal{}, fmt.Errorf("computing [%v / %v]: %w", d, e, ErrDivisionByZero)
	}
	if !mode.valid() {
		return Decimal{}, fmt.Errorf("computing [%v / %v]: %w", d, e, ErrInvalidMode)
	}
	scale = min(max(scale, scaleMin), MaxScale)

	// Special case: zero dividend
	if d.IsZero() {
		return newUnsafe(false, 0, scale), nil
	}

	f, err := d.quoRoundFint(e, scale, mode)
	if err != nil {
		f, err = d.quoRoundBint(e, scale, mode)
		if err != nil {
			return Decimal{}, fmt.Errorf("computing [%v / %v]: %w", d, e, err)
		}
	}
	return f, nil
}

// quoRoundFint computes QuoRound using uint64 arithmetic.
// The quotient is x = a / b, where a and b are the coefficients aligned to
// a common scale, so that the integer part q0 and the remainder r0 are exact.
func (d Decimal) quoRoundFint(e Decimal, scale int, mode RoundingMode) (Decimal, error) {
	acoef := d.coef
	bcoef := e.coef
	neg := d.neg != e.neg

	// Alignment
	var ok bool
	switch {
	case e.Scale() > d.Scale():
		acoef, ok = acoef.lsh(e.Scale() - d.Scale())
	case d.Scale() > e.Scale():
		bcoef, ok = bcoef.lsh(d.Scale() - e.Scale())
	default:
		ok = true
	}
	if !ok {
		return Decimal{}, ErrOverflow
	}

	// Integer part, and the scale that fits next to it
	q0, r0, _ := acoef.quoRem(bcoef)
	scale = min(scale, MaxPrec-q0.prec())

	// Fractional digits: n = q0 * 10^scale + ⌊r0 * 10^scale / b⌋.
	// The product is 128-bit, and the quotient fits uint64 because r0 < b.
	hi, lo := bits.Mul64(uint64(r0), uint64(pow10[scale]))
	q1, r1 := bits.Div64(hi, lo, uint64(bcoef))
	n := q0*pow10[scale] + fint(q1)

	if mode.inc(neg, n, fint(r1), bcoef) {
		n++
	}
	// Rounding carried into a 20th digit: n is exactly 10^19
	if n > maxCoef {
		if scale == 0 {
			return Decimal{}, ErrOverflow
		}
		n /= 10
		scale--
	}
	return newSafe(neg, n, scale)
}

// quoRoundBint computes QuoRound using *big.Int arithmetic.
func (d Decimal) quoRoundBint(e Decimal, scale int, mode RoundingMode) (Decimal, error) {
	acoef := getBint()
	defer putBint(acoef)
	bcoef := getBint()
	defer putBint(bcoef)
	qcoef := getBint()
	defer putBint(qcoef)
	rcoef := getBint()
	defer putBint(rcoef)
	tcoef := getBint()
	defer putBint(tcoef)

	acoef.setFint(d.coef)
	bcoef.setFint(e.coef)
	neg := d.neg != e.neg

	// Alignment
	switch {
	case e.Scale() > d.Scale():
		acoef.lsh(acoef, e.Scale()-d.Scale())
	case d.Scale() > e.Scale():
		bcoef.lsh(bcoef, d.Scale()-e.Scale())
	}

	// Integer part, and the scale that fits next to it
	qcoef.quoRem(acoef, bcoef, rcoef)
	digits := qcoef.prec()
	if digits > MaxPrec {
		return Decimal{}, fmt.Errorf("%w: the integer part of a %T can have at most %v digits, but it has %v digits", ErrOverflow, Decimal{}, MaxPrec, digits)
	}
	scale = min(scale, MaxPrec-digits)

	// Fractional digits: n = q0 * 10^scale + ⌊r0 * 10^scale / b⌋,
	// and acoef becomes the final remainder.
	rcoef.lsh(rcoef, scale)
	tcoef.quoRem(rcoef, bcoef, acoef)
	qcoef.lsh(qcoef, scale)
	qcoef.add(qcoef, tcoef)

	if mode.incBint(neg, qcoef, acoef, bcoef) {
		qcoef.inc(qcoef)
	}
	// Rounding carried into a 20th digit: n is exactly 10^19
	if qcoef.hasPrec(MaxPrec + 1) {
		if scale == 0 {
			return Decimal{}, fmt.Errorf("%w: the integer part of a %T can have at most %v digits, but it has %v digits", ErrOverflow, Decimal{}, MaxPrec, MaxPrec+1)
		}
		qcoef.quo(qcoef, bpow10[1])
		scale--
	}
	return newSafe(neg, qcoef.fint(), scale)
}
