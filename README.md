# decimal

Exact decimal arithmetic for prices, quantities and money: parsing and
printing decimal text, the four operations, comparison, explicit rounding
modes and rounding to an instrument step. No dependencies beyond the
standard library, no heap allocations on the fast path, no panics on bad
data (only `MustParse`, and `Round` with an unknown mode, panic).

The representation and the arithmetic core are derived from
[govalues/decimal](https://github.com/govalues/decimal) v0.1.36 (MIT, see
LICENSE); NOTICE.md lists what was taken and what differs.

A value has up to 19 significant digits and up to 19 digits after the
decimal point. An exact result that needs more digits is rounded half to
even; an integer part that does not fit is an error.

## API overview

- Construction: `Parse`, `MustParse`, `New(coef, scale)`, `NewFromFloat64`.
- Arithmetic (all return an error instead of panicking): `Add`, `Sub`, `Mul`,
  `Quo`, `QuoRem`, and `QuoRound(e, scale, mode)`, the exact quotient rounded
  to `scale` digits by a rounding mode.
- `Min(a, b)`, `Max(a, b)`.
- Comparison and predicates: `Cmp`, `Equal`, `Sign`, `IsZero`, `IsPos`,
  `IsNeg`, `IsInt`, `Neg`, `Abs`, `Scale`, `Prec`.
- Rounding: `Round(scale, mode)`, `RoundStep(step, mode)`, `IsMultipleOf(step)`,
  with modes `Floor`, `Ceil`, `Down`, `Up`, `HalfUp`, `HalfEven`.
- Scale: `Rescale(scale)` pads with zeros (or drops zeros), `Trim()`.
- Output: `String`, `Float64`, `Int64(scale)`.
- Encoding: `MarshalJSON` writes a bare JSON number (`{"price":65000.10}`);
  `UnmarshalJSON` reads a JSON number or a string and rejects `null`;
  `MarshalText`, `UnmarshalText` and `AppendText` (`encoding.TextAppender`)
  use the `String`/`Parse` forms.
- Errors, for `errors.Is`: `ErrOverflow`, `ErrDivisionByZero`, `ErrSyntax`,
  `ErrInvalidStep`, `ErrInvalidMode`, `ErrInexact`.

## Example

```go
price := decimal.MustParse("65000.17")
qty := decimal.MustParse("0.012")

notional, err := price.Mul(qty) // exact: 780.00204
if err != nil {
	return err
}

tick := decimal.MustParse("0.5")
limit, err := price.RoundStep(tick, decimal.Floor) // 65000.0
if err != nil {
	return err
}

lot := decimal.MustParse("0.001")
ok := qty.IsMultipleOf(lot) // true

fmt.Println(notional, limit, ok, notional.Round(2, decimal.HalfEven))
// 780.00204 65000.0 true 780.00
```

`RoundStep` returns the result at the scale of the step (`1.5` with step
`0.01` is `1.50`) and decides the rounding from the exact remainder, so it is
safe for values with 19 digits. Scale is kept by arithmetic: `1.50` and `1.5`
are `Equal` but print differently; use `Trim` or `Rescale` to normalise.
