// Package decimal implements exact fixed-point decimal arithmetic for
// prices and quantities.
//
// It is a minimal subset of github.com/govalues/decimal v0.1.36 (MIT),
// see LICENSE and NOTICE.md: the representation, parsing and printing,
// Add, Sub, Mul, Quo, QuoRem and comparison.
//
// A [Decimal] holds a sign, a coefficient of at most 19 decimal digits and
// a scale (the number of digits after the decimal point) of at most 19.
// Scale is preserved: 1.50 and 1.5 are [Decimal.Equal] but print differently.
//
// When a result needs more than 19 digits, it is rounded to fit using
// half-to-even (banker's) rounding: to 19 significant digits if there is an
// integer part, otherwise to 19 digits after the decimal point.
// A result whose integer part has more than 19 digits is an error.
//
// Operations never panic on bad input: they return errors satisfying
// errors.Is with [ErrOverflow], [ErrDivisionByZero] or [ErrSyntax].
// Only [MustParse] panics.
package decimal
