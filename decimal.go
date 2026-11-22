// Derived from github.com/govalues/decimal v0.1.36 (MIT), see LICENSE.
package decimal

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"unsafe"
)

// Decimal represents a finite floating-point decimal number.
// Its zero value corresponds to the numeric value of 0.
// Decimal is designed to be safe for concurrent use by multiple goroutines.
type Decimal struct {
	neg   bool // indicates whether the decimal is negative
	scale int8 // position of the floating decimal point
	coef  fint // numeric value without decimal point
}

const (
	MaxPrec  = 19      // MaxPrec is the maximum length of the coefficient in decimal digits.
	scaleMin = 0       // scaleMin is the minimum number of digits after the decimal point.
	MaxScale = 19      // MaxScale is the maximum number of digits after the decimal point.
	maxCoef  = maxFint // maxCoef is the maximum absolute value of the coefficient, which is equal to (10^MaxPrec - 1).
)

// Exported sentinel errors. Every error returned by this package
// satisfies errors.Is with one of them.
var (
	ErrOverflow       = errors.New("decimal overflow")
	ErrDivisionByZero = errors.New("division by zero")
	ErrSyntax         = errors.New("invalid decimal")
)

var (
	Zero = Decimal{}        // Zero represents the decimal value of 0. For comparison purposes, use the IsZero method.
	One  = Decimal{coef: 1} // One represents the decimal value of 1.

	errScaleRange      = fmt.Errorf("%w: scale out of range", ErrOverflow)
	errInexactDivision = errors.New("inexact division") // internal: makes quoFint fall back to quoBint
)

// newUnsafe creates a new decimal without checking the scale and coefficient.
// Use it only if you are absolutely sure that the arguments are valid.
func newUnsafe(neg bool, coef fint, scale int) Decimal {
	if coef == 0 {
		neg = false
	}
	//nolint:gosec
	return Decimal{neg: neg, coef: coef, scale: int8(scale)}
}

// newSafe creates a new decimal and checks the scale and coefficient.
func newSafe(neg bool, coef fint, scale int) (Decimal, error) {
	switch {
	case scale < scaleMin || scale > MaxScale:
		return Decimal{}, errScaleRange
	case coef > maxCoef:
		return Decimal{}, ErrOverflow
	}
	return newUnsafe(neg, coef, scale), nil
}

// newFromFint creates a new decimal from a uint64 coefficient.
// This method does not use overflowError to return descriptive errors,
// as it must be as fast as possible.
func newFromFint(neg bool, coef fint, scale, minScale int) (Decimal, error) {
	var ok bool
	// Scale normalization
	switch {
	case scale < minScale:
		coef, ok = coef.lsh(minScale - scale)
		if !ok {
			return Decimal{}, ErrOverflow
		}
		scale = minScale
	case scale > MaxScale:
		coef = coef.rshHalfEven(scale - MaxScale)
		scale = MaxScale
	}
	return newSafe(neg, coef, scale)
}

// newFromBint creates a new decimal from a *big.Int coefficient.
// This method uses overflowError to return descriptive errors.
func newFromBint(neg bool, coef *bint, scale, minScale int) (Decimal, error) {
	// Overflow validation
	prec := coef.prec()
	if prec-scale > MaxPrec-minScale {
		return Decimal{}, overflowError(prec, scale, minScale)
	}
	// Scale normalization
	switch {
	case scale < minScale:
		coef.lsh(coef, minScale-scale)
		scale = minScale
	case scale >= prec && scale > MaxScale: // no integer part
		coef.rshHalfEven(coef, scale-MaxScale)
		scale = MaxScale
	case prec > scale && prec > MaxPrec: // there is an integer part
		coef.rshHalfEven(coef, prec-MaxPrec)
		scale = MaxPrec - prec + scale
	}
	// Handling the rare case when rshHalfEven rounded
	// a 19-digit coefficient to a 20-digit coefficient.
	if coef.hasPrec(MaxPrec + 1) {
		return newFromBint(neg, coef, scale, minScale)
	}
	return newSafe(neg, coef.fint(), scale)
}

func overflowError(gotPrec, gotScale, wantScale int) error {
	maxDigits := MaxPrec - wantScale
	gotDigits := gotPrec - gotScale
	switch wantScale {
	case 0:
		return fmt.Errorf("%w: the integer part of a %T can have at most %v digits, but it has %v digits", ErrOverflow, Decimal{}, maxDigits, gotDigits)
	default:
		return fmt.Errorf("%w: with %v significant digits after the decimal point, the integer part of a %T can have at most %v digits, but it has %v digits", ErrOverflow, wantScale, Decimal{}, maxDigits, gotDigits)
	}
}

// New returns a decimal equal to value / 10^scale.
// New keeps trailing zeros in the fractional part to preserve scale.
//
// New returns an error if the scale is negative or greater than [MaxScale],
// or if |value| has more than [MaxPrec] digits.
func New(value int64, scale int) (Decimal, error) {
	var coef fint
	var neg bool
	if value >= 0 {
		neg = false
		coef = fint(value)
	} else {
		neg = true
		if value == math.MinInt64 {
			coef = fint(math.MaxInt64) + 1
		} else {
			coef = fint(-value)
		}
	}
	return newSafe(neg, coef, scale)
}

// MustParse is like [Parse] but panics if the string cannot be parsed.
// It simplifies safe initialization of global variables holding decimals.
func MustParse(s string) Decimal {
	d, err := Parse(s)
	if err != nil {
		panic(fmt.Sprintf("Parse(%q) failed: %v", s, err))
	}
	return d
}

// Parse converts a string to a (possibly rounded) decimal.
// The input string must be in one of the following formats:
//
//	1.234
//	-1234
//	+0.000001234
//	1.83e5
//	0.22e-9
//
// The formal EBNF grammar for the supported format is as follows:
//
//	sign           ::= '+' | '-'
//	digits         ::= { '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9' }
//	significand    ::= digits '.' digits | '.' digits | digits '.' | digits
//	exponent       ::= ('e' | 'E') [sign] digits
//	numeric-string ::= [sign] significand [exponent]
//
// Parse removes leading and trailing whitespace (see [strings.TrimSpace])
// before parsing.
// Parse removes leading zeros from the integer part of the input string,
// but tries to maintain trailing zeros in the fractional part to preserve scale.
//
// Parse returns an error satisfying errors.Is(err, [ErrSyntax]) if:
//   - the string contains whitespace other than at its ends;
//   - the string is longer than 330 bytes;
//   - the exponent is less than -330 or greater than 330;
//   - the string does not represent a valid decimal number.
//
// Parse returns an error satisfying errors.Is(err, [ErrOverflow]) if
// the integer part of the result has more than [MaxPrec] digits.
func Parse(s string) (Decimal, error) {
	s = strings.TrimSpace(s)
	text := unsafe.Slice(unsafe.StringData(s), len(s))
	return parse(text)
}

func parse(text []byte) (Decimal, error) {
	if len(text) > 330 {
		return Decimal{}, fmt.Errorf("parsing decimal: %w", ErrSyntax)
	}
	d, err := parseFint(text, 0)
	if err != nil {
		d, err = parseBint(text, 0)
		if err != nil {
			return Decimal{}, fmt.Errorf("parsing decimal: %w", err)
		}
	}
	return d, nil
}

// parseFint parses a decimal string using uint64 arithmetic.
// parseFint does not support exponential notation to make it as fast as possible.
//
//nolint:gocyclo
func parseFint(text []byte, minScale int) (Decimal, error) {
	var pos int
	width := len(text)

	// Sign
	var neg bool
	switch {
	case pos == width:
		// skip
	case text[pos] == '-':
		neg = true
		pos++
	case text[pos] == '+':
		pos++
	}

	// Coefficient
	var coef fint
	var scale int
	var hasCoef, ok bool

	// Integer
	for pos < width && text[pos] >= '0' && text[pos] <= '9' {
		coef, ok = coef.fsa(1, text[pos]-'0')
		if !ok {
			return Decimal{}, ErrOverflow
		}
		pos++
		hasCoef = true
	}

	// Fraction
	if pos < width && text[pos] == '.' {
		pos++
		for pos < width && text[pos] >= '0' && text[pos] <= '9' {
			coef, ok = coef.fsa(1, text[pos]-'0')
			if !ok {
				return Decimal{}, ErrOverflow
			}
			pos++
			scale++
			hasCoef = true
		}
	}

	if pos != width {
		return Decimal{}, fmt.Errorf("%w: unexpected character %q", ErrSyntax, text[pos])
	}
	if !hasCoef {
		return Decimal{}, fmt.Errorf("%w: no coefficient", ErrSyntax)
	}
	return newFromFint(neg, coef, scale, minScale)
}

// parseBint parses a decimal string using *big.Int arithmetic.
// parseBint supports exponential notation.
//
//nolint:gocyclo
func parseBint(text []byte, minScale int) (Decimal, error) {
	var pos int
	width := len(text)

	// Sign
	var neg bool
	switch {
	case pos == width:
		// skip
	case text[pos] == '-':
		neg = true
		pos++
	case text[pos] == '+':
		pos++
	}

	// Coefficient
	bcoef := getBint()
	defer putBint(bcoef)
	var fcoef fint
	var shift, scale int
	var hasCoef, ok bool

	bcoef.setFint(0)

	// Algorithm:
	// 	1. Add as many digits as possible to the uint64 coefficient (fast).
	// 	2. Once the uint64 coefficient has reached its maximum value,
	//     add it to the *big.Int coefficient (slow).
	// 	3. Repeat until all digits are processed.

	// Integer
	for pos < width && text[pos] >= '0' && text[pos] <= '9' {
		fcoef, ok = fcoef.fsa(1, text[pos]-'0')
		if !ok {
			return Decimal{}, ErrOverflow // Should never happen
		}
		pos++
		shift++
		hasCoef = true
		if fcoef.hasPrec(MaxPrec) {
			bcoef.fsa(bcoef, shift, fcoef)
			fcoef, shift = 0, 0
		}
	}

	// Fraction
	if pos < width && text[pos] == '.' {
		pos++
		for pos < width && text[pos] >= '0' && text[pos] <= '9' {
			fcoef, ok = fcoef.fsa(1, text[pos]-'0')
			if !ok {
				return Decimal{}, ErrOverflow // Should never happen
			}
			pos++
			scale++
			shift++
			hasCoef = true
			if fcoef.hasPrec(MaxPrec) {
				bcoef.fsa(bcoef, shift, fcoef)
				fcoef, shift = 0, 0
			}
		}
	}
	if shift > 0 {
		bcoef.fsa(bcoef, shift, fcoef)
	}

	// Exponent
	var exp int
	var eneg, hasExp, hasE bool
	if pos < width && (text[pos] == 'e' || text[pos] == 'E') {
		pos++
		hasE = true
		// Sign
		switch {
		case pos == width:
			// skip
		case text[pos] == '-':
			eneg = true
			pos++
		case text[pos] == '+':
			pos++
		}
		// Integer
		for pos < width && text[pos] >= '0' && text[pos] <= '9' {
			exp = exp*10 + int(text[pos]-'0')
			if exp > 330 {
				return Decimal{}, ErrSyntax
			}
			pos++
			hasExp = true
		}
	}

	if pos != width {
		return Decimal{}, fmt.Errorf("%w: unexpected character %q", ErrSyntax, text[pos])
	}
	if !hasCoef {
		return Decimal{}, fmt.Errorf("%w: no coefficient", ErrSyntax)
	}
	if hasE && !hasExp {
		return Decimal{}, fmt.Errorf("%w: no exponent", ErrSyntax)
	}

	if eneg {
		scale = scale + exp
	} else {
		scale = scale - exp
	}

	return newFromBint(neg, bcoef, scale, minScale)
}

// String implements the [fmt.Stringer] interface and returns
// a string representation of the decimal.
// The returned string does not use scientific or engineering notation and is
// formatted according to the following formal EBNF grammar:
//
//	sign           ::= '-'
//	digits         ::= { '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9' }
//	significand    ::= digits '.' digits | digits
//	numeric-string ::= [sign] significand
//
// [fmt.Stringer]: https://pkg.go.dev/fmt#Stringer
func (d Decimal) String() string {
	text := make([]byte, 0, 24)
	return string(d.append(text))
}

// append appends a string representation of the decimal to the byte slice.
func (d Decimal) append(text []byte) []byte {
	var buf [24]byte
	pos := len(buf) - 1
	coef := d.coef
	scale := d.Scale()

	// Coefficient
	for {
		buf[pos] = byte(coef%10) + '0'
		pos--
		coef /= 10
		if scale > 0 {
			scale--
			// Decimal point
			if scale == 0 {
				buf[pos] = '.'
				pos--
				// Leading 0
				if coef == 0 {
					buf[pos] = '0'
					pos--
				}
			}
		}
		if coef == 0 && scale == 0 {
			break
		}
	}

	// Sign
	if d.IsNeg() {
		buf[pos] = '-'
		pos--
	}

	return append(text, buf[pos+1:]...)
}

// Prec returns the number of digits in the coefficient.
func (d Decimal) Prec() int {
	return d.coef.prec()
}

// Scale returns the number of digits after the decimal point.
func (d Decimal) Scale() int {
	return int(d.scale)
}

// minScale returns the smallest scale that the decimal can be rescaled to
// without rounding.
func (d Decimal) minScale() int {
	// Special case: zero
	if d.IsZero() {
		return scaleMin
	}
	// General case
	dcoef := d.coef
	return max(scaleMin, d.Scale()-dcoef.ntz())
}

// IsInt returns true if there are no significant digits after the decimal point.
func (d Decimal) IsInt() bool {
	return d.Scale() == 0 || d.coef%pow10[d.Scale()] == 0
}

// trunc returns a decimal truncated to the specified number of digits
// after the decimal point using [rounding toward zero].
// If the given scale is negative, it is redefined to zero.
//
// [rounding toward zero]: https://en.wikipedia.org/wiki/Rounding#Rounding_toward_zero
func (d Decimal) trunc(scale int) Decimal {
	scale = max(scale, scaleMin)
	if scale >= d.Scale() {
		return d
	}
	coef := d.coef
	coef = coef.rshDown(d.Scale() - scale)
	return newUnsafe(d.IsNeg(), coef, scale)
}

// trim returns a decimal with trailing zeros removed up to the given number of
// digits after the decimal point.
// If the given scale is negative, it is redefined to zero.
func (d Decimal) trim(scale int) Decimal {
	if d.Scale() <= scale {
		return d
	}
	scale = max(scale, d.minScale())
	return d.trunc(scale)
}

// Neg returns a decimal with the opposite sign.
func (d Decimal) Neg() Decimal {
	return newUnsafe(!d.IsNeg(), d.coef, d.Scale())
}

// Abs returns the absolute value of the decimal.
func (d Decimal) Abs() Decimal {
	return newUnsafe(false, d.coef, d.Scale())
}

// Sign returns:
//
//	-1 if d < 0
//	 0 if d = 0
//	+1 if d > 0
func (d Decimal) Sign() int {
	switch {
	case d.neg:
		return -1
	case d.coef == 0:
		return 0
	}
	return 1
}

// IsPos returns:
//
//	true  if d > 0
//	false otherwise
func (d Decimal) IsPos() bool {
	return d.coef != 0 && !d.neg
}

// IsNeg returns:
//
//	true  if d < 0
//	false otherwise
func (d Decimal) IsNeg() bool {
	return d.neg
}

// IsZero returns:
//
//	true  if d = 0
//	false otherwise
func (d Decimal) IsZero() bool {
	return d.coef == 0
}
