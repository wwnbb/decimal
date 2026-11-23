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
// rounding uses the same mode at that scale. A result that rounds to zero is
// zero, never negative.
//
// Rounding never carries out of the digit budget (to 10^19), so no
// renormalization is needed. Let k be the number of digits of the integer part
// q0, s = MaxPrec-k the scale that fits next to it, and a / b the aligned
// operands. With a smaller scale the rounded n is at most 10^(k+scale) < 10^19.
// With s, a carry needs q0 = 10^k-1 and ⌊a*10^s/b⌋ = 10^19-1; only a nonzero
// remainder can round up, so j = 10^k*b - a satisfies j >= 1 and, strictly,
// j*10^s < b.
// Alignment shifts only one operand, by 10^t:
//   - b is shifted or nothing is: a is a raw coefficient, so a < 10^19 and
//     10^k*b = a+j <= 10^19-1+j,
//     while j*10^s < b gives 10^k*b > j*10^19; hence j*10^19 < 10^19-1+j,
//     that is j < 1, a contradiction.
//   - a = d*10^t is shifted and t <= k: j = j'*10^t with j' >= 1 and
//     d = 10^(k-t)*b - j' < 10^19, while b > j'*10^(s+t) gives
//     10^(k-t)*b > j'*10^19: the same contradiction.
//   - a is shifted and t > k: j is a multiple of 10^k, so b > j*10^s >= 10^19,
//     but b < 10^19.
//
// The model check of FuzzQuoRound asserts that the rounded coefficient stays
// below 10^19.
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
	return newSafe(neg, n, scale)
}

// quoRoundBint computes QuoRound using *big.Int arithmetic. Like
// quoRoundFint it relies on rounding never carrying to 10^19 (see QuoRound);
// newSafe still refuses a coefficient above maxCoef.
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
	return newSafe(neg, qcoef.fint(), scale)
}
