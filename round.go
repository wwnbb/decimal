package decimal

import (
	"errors"
	"fmt"
)

// RoundingMode selects how a value is rounded to an integer multiple of a
// unit (a power of ten for [Decimal.Round], a step for [Decimal.RoundStep]).
type RoundingMode uint8

const (
	Floor    RoundingMode = iota // toward −∞
	Ceil                         // toward +∞
	Down                         // toward zero
	Up                           // away from zero
	HalfUp                       // nearest, halves away from zero
	HalfEven                     // nearest, halves to even
)

var (
	// ErrInvalidStep is returned by [Decimal.RoundStep] when the step is not positive.
	ErrInvalidStep = errors.New("step must be positive")
	// ErrInvalidMode is returned when a [RoundingMode] is not one of the constants.
	ErrInvalidMode = errors.New("invalid rounding mode")
	// ErrInexact is returned by [Decimal.Rescale] when digits would be lost.
	ErrInexact = errors.New("rescale would lose digits")
)

// String returns the name of the mode.
func (m RoundingMode) String() string {
	switch m {
	case Floor:
		return "Floor"
	case Ceil:
		return "Ceil"
	case Down:
		return "Down"
	case Up:
		return "Up"
	case HalfUp:
		return "HalfUp"
	case HalfEven:
		return "HalfEven"
	}
	return fmt.Sprintf("RoundingMode(%d)", uint8(m))
}

func (m RoundingMode) valid() bool {
	return m <= HalfEven
}

// inc reports whether the magnitude q of a truncated quotient x / y must be
// increased by one, where r = x - q*y is the remainder, 0 <= r < y, and neg
// is the sign of x.
func (m RoundingMode) inc(neg bool, q, r, y fint) bool {
	if r == 0 {
		return false
	}
	switch m {
	case Floor:
		return neg
	case Ceil:
		return !neg
	case Down:
		return false
	case Up:
		return true
	case HalfUp:
		return r >= y-r
	}
	return r > y-r || (r == y-r && q.isOdd()) // HalfEven
}

// incBint is inc for *big.Int operands.
func (m RoundingMode) incBint(neg bool, q, r, y *bint) bool {
	if r.sign() == 0 {
		return false
	}
	switch m {
	case Floor:
		return neg
	case Ceil:
		return !neg
	case Down:
		return false
	case Up:
		return true
	}
	twice := getBint()
	defer putBint(twice)
	twice.dbl(r)
	c := twice.cmp(y)
	if m == HalfUp {
		return c >= 0
	}
	return c > 0 || (c == 0 && q.isOdd()) // HalfEven
}

// Round returns d rounded to the given number of digits after the decimal
// point using the mode. The scale is clamped to [0, MaxScale]. If the scale
// of d is already not greater than scale, d is returned unchanged: Round
// never pads with zeros. A result that rounds to zero is zero, never negative.
//
// Round cannot overflow: it only runs when at least one digit is dropped,
// so the truncated coefficient is at most maxFint/10 and adding one to it
// still fits.
//
// Round panics if the mode is not one of the constants of [RoundingMode].
func (d Decimal) Round(scale int, mode RoundingMode) Decimal {
	if !mode.valid() {
		panic(fmt.Sprintf("decimal: %v", ErrInvalidMode))
	}
	scale = min(max(scale, scaleMin), MaxScale)
	if d.Scale() <= scale {
		return d
	}
	y := pow10[d.Scale()-scale] // shift is in 1..MaxScale
	q, r, _ := d.coef.quoRem(y)
	if mode.inc(d.neg, q, r, y) {
		q++
	}
	return newUnsafe(d.neg, q, scale)
}

// Rescale returns d with exactly the given scale and the same value.
// It pads with zeros if scale is greater than the scale of d, which returns
// an error matching [ErrOverflow] if the coefficient would exceed 19 digits.
// If scale is smaller, the dropped digits must all be zeros, otherwise the
// error matches [ErrInexact]. A scale outside [0, MaxScale] matches [ErrOverflow].
func (d Decimal) Rescale(scale int) (Decimal, error) {
	switch {
	case scale < scaleMin || scale > MaxScale:
		return Decimal{}, fmt.Errorf("rescaling %v to %v: %w", d, scale, errScaleRange)
	case scale == d.Scale():
		return d, nil
	case scale > d.Scale():
		coef, ok := d.coef.lsh(scale - d.Scale())
		if !ok {
			return Decimal{}, fmt.Errorf("rescaling %v to %v: %w", d, scale, ErrOverflow)
		}
		return newUnsafe(d.neg, coef, scale), nil
	}
	y := pow10[d.Scale()-scale]
	q, r, _ := d.coef.quoRem(y)
	if r != 0 {
		return Decimal{}, fmt.Errorf("rescaling %v to %v: %w", d, scale, ErrInexact)
	}
	return newUnsafe(d.neg, q, scale), nil
}

// Trim returns d without trailing zeros in the fractional part:
// 1.500 becomes 1.5 and 100.00 becomes 100.
func (d Decimal) Trim() Decimal {
	return d.trim(0)
}

// RoundStep returns n*step, where n is d / step rounded to an integer using
// the mode. The computation is exact: the rounding is decided from the exact
// integer quotient and remainder, never from a rounded quotient.
//
// The result has the scale of step, so 1.5 with step 0.01 gives 1.50.
//
// RoundStep returns an error matching [ErrInvalidStep] if step is not
// positive, [ErrInvalidMode] if the mode is unknown, and [ErrOverflow] if
// n*step needs more than 19 digits.
func (d Decimal) RoundStep(step Decimal, mode RoundingMode) (Decimal, error) {
	if !step.IsPos() {
		return Decimal{}, fmt.Errorf("rounding %v to step %v: %w", d, step, ErrInvalidStep)
	}
	if !mode.valid() {
		return Decimal{}, fmt.Errorf("rounding %v to step %v: %w", d, step, ErrInvalidMode)
	}
	f, err := d.roundStepFint(step, mode)
	if err != nil {
		f, err = d.roundStepBint(step, mode)
		if err != nil {
			return Decimal{}, fmt.Errorf("rounding %v to step %v: %w", d, step, err)
		}
	}
	return f, nil
}

// roundStepFint computes RoundStep using uint64 arithmetic.
func (d Decimal) roundStepFint(step Decimal, mode RoundingMode) (Decimal, error) {
	dcoef := d.coef
	scoef := step.coef

	// Alignment
	var ok bool
	switch {
	case d.Scale() > step.Scale():
		scoef, ok = scoef.lsh(d.Scale() - step.Scale())
	case d.Scale() < step.Scale():
		dcoef, ok = dcoef.lsh(step.Scale() - d.Scale())
	default:
		ok = true
	}
	if !ok {
		return Decimal{}, ErrOverflow
	}

	// Compute n = round(d / step)
	q, r, _ := dcoef.quoRem(scoef)
	if mode.inc(d.neg, q, r, scoef) {
		q++
	}

	// Compute n * step
	coef, ok := q.mul(step.coef)
	if !ok {
		return Decimal{}, ErrOverflow
	}
	return newSafe(d.neg, coef, step.Scale())
}

// roundStepBint computes RoundStep using *big.Int arithmetic.
func (d Decimal) roundStepBint(step Decimal, mode RoundingMode) (Decimal, error) {
	dcoef := getBint()
	defer putBint(dcoef)
	scoef := getBint()
	defer putBint(scoef)
	qcoef := getBint()
	defer putBint(qcoef)
	rcoef := getBint()
	defer putBint(rcoef)

	dcoef.setFint(d.coef)
	scoef.setFint(step.coef)

	// Alignment
	switch {
	case d.Scale() > step.Scale():
		scoef.lsh(scoef, d.Scale()-step.Scale())
	case d.Scale() < step.Scale():
		dcoef.lsh(dcoef, step.Scale()-d.Scale())
	}

	// Compute n = round(d / step)
	qcoef.quoRem(dcoef, scoef, rcoef)
	if mode.incBint(d.neg, qcoef, rcoef, scoef) {
		qcoef.inc(qcoef)
	}

	// Compute n * step
	rcoef.setFint(step.coef)
	qcoef.mul(qcoef, rcoef)
	if qcoef.hasPrec(MaxPrec + 1) {
		return Decimal{}, fmt.Errorf("%w: the result needs more than %v digits", ErrOverflow, MaxPrec)
	}
	return newSafe(d.neg, qcoef.fint(), step.Scale())
}

// IsMultipleOf reports whether d is an integer multiple of step, that is
// whether d / step is an integer. It returns false if step is not positive.
// The check is exact and never fails: it looks only at the remainder, so a
// quotient with more than 19 digits is not a problem.
func (d Decimal) IsMultipleOf(step Decimal) bool {
	if !step.IsPos() {
		return false
	}
	dcoef := d.coef
	scoef := step.coef

	// Alignment
	var ok bool
	switch {
	case d.Scale() > step.Scale():
		scoef, ok = scoef.lsh(d.Scale() - step.Scale())
	case d.Scale() < step.Scale():
		dcoef, ok = dcoef.lsh(step.Scale() - d.Scale())
	default:
		ok = true
	}
	if ok {
		return dcoef%scoef == 0
	}

	// The aligned coefficient does not fit uint64, use *big.Int.
	bd := getBint()
	defer putBint(bd)
	bs := getBint()
	defer putBint(bs)
	bq := getBint()
	defer putBint(bq)
	br := getBint()
	defer putBint(br)
	bd.setFint(d.coef)
	bs.setFint(step.coef)
	if d.Scale() > step.Scale() {
		bs.lsh(bs, d.Scale()-step.Scale())
	} else {
		bd.lsh(bd, step.Scale()-d.Scale())
	}
	bq.quoRem(bd, bs, br)
	return br.sign() == 0
}
