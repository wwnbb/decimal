package decimal

import (
	"fmt"
	"math"
	"strconv"
)

// NewFromFloat64 converts a float to a decimal using the shortest decimal
// string that round-trips the float (strconv.FormatFloat with 'g' and
// precision -1), then [Parse]. So 0.1 gives 0.1, not 0.1000000000000000055.
//
// Digits beyond the 19th after the decimal point are rounded half to even,
// exactly as [Parse] does, so a tiny nonzero float such as 1e-25 becomes
// 0.0000000000000000000.
//
// NewFromFloat64 returns an error matching [ErrSyntax] for NaN and ±Inf and
// [ErrOverflow] if the integer part has more than [MaxPrec] digits.
func NewFromFloat64(f float64) (Decimal, error) {
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return Decimal{}, fmt.Errorf("converting float: %w: special value %v", ErrSyntax, f)
	}
	var buf [32]byte
	text := strconv.AppendFloat(buf[:0], f, 'g', -1, 64)
	d, err := parse(text)
	if err != nil {
		return Decimal{}, fmt.Errorf("converting float: %w", err)
	}
	return d, nil
}

// Float64 returns the float nearest to d (strconv.ParseFloat rounds half to
// even). The conversion may lose digits, as a float64 has fewer than 19
// digits of precision; ok is false only if the conversion fails, which does
// not happen for any valid decimal.
func (d Decimal) Float64() (f float64, ok bool) {
	f, err := strconv.ParseFloat(d.String(), 64)
	if err != nil {
		return 0, false
	}
	return f, true
}

// Int64 returns d * 10^scale as an integer, so that 1.25 with scale 2 is 125.
// ok is false if the result is not an integer (digits would be lost), does
// not fit int64, or the scale is outside [0, MaxScale].
func (d Decimal) Int64(scale int) (n int64, ok bool) {
	if scale < scaleMin || scale > MaxScale {
		return 0, false
	}
	coef := d.coef
	if scale >= d.Scale() {
		coef, ok = coef.lsh(scale - d.Scale())
		if !ok {
			return 0, false
		}
	} else {
		var r fint
		coef, r, _ = coef.quoRem(pow10[d.Scale()-scale])
		if r != 0 {
			return 0, false
		}
	}
	if d.neg {
		if coef > fint(math.MaxInt64)+1 {
			return 0, false
		}
		return int64(-coef), true //nolint:gosec
	}
	if coef > math.MaxInt64 {
		return 0, false
	}
	return int64(coef), true //nolint:gosec
}
