# NOTICE

This package is derived from [github.com/govalues/decimal](https://github.com/govalues/decimal)
(MIT, see `LICENSE`).

- Tag: `v0.1.36`
- Commit: `377e9df6035e8b27bc4a31d1af6e23e29ca4adad`

Code was copied function by function; invariants and comments were kept.
Files with ported code start with `// Derived from github.com/govalues/decimal v0.1.36 (MIT), see LICENSE.`
`fuzz_test.go` and `doc.go` are new.

## Upstream function → here

| Upstream (`decimal.go` / `integer.go`) | Here |
| --- | --- |
| `type Decimal`, `MaxPrec`, `MaxScale`, `maxCoef`, `Zero`, `One` | `decimal.go` |
| `errDecimalOverflow`, `errInvalidDecimal`, `errDivisionByZero` | `ErrOverflow`, `ErrSyntax`, `ErrDivisionByZero` in `decimal.go` |
| `errScaleRange` | `errScaleRange` in `decimal.go`, wraps `ErrOverflow` |
| `errInexactDivision` | `errInexactDivision` in `decimal.go` (internal) |
| `newUnsafe`, `newSafe`, `newFromFint`, `newFromBint`, `overflowError` | same names, `decimal.go` |
| `New` | `New`, `decimal.go` |
| `MustParse`, `Parse`, `parse`, `parseExact`, `parseFint`, `parseBint` | `MustParse`, `Parse` (adds `TrimSpace`), `parse` (`parseExact` with scale 0), `parseFint`, `parseBint`, `decimal.go` |
| `String`, `bytes`, `append` | `String` (merges `bytes`), `append`, `decimal.go` |
| `Prec`, `Scale`, `IsInt`, `Neg`, `Abs`, `Sign`, `IsPos`, `IsNeg`, `IsZero` | same names, `decimal.go` |
| `MinScale` (method), `Trunc`, `Trim` | unexported `minScale`, `trunc`, `trim`, `decimal.go` (used by `Quo`) |
| `Mul`, `MulExact`, `mulFint`, `mulBint` | `Mul` (`MulExact` with scale 0), `mulFint`, `mulBint`, `arith.go` |
| `Add`, `AddExact`, `addFint`, `addBint` | `Add` (`AddExact` with scale 0), `addFint`, `addBint`, `arith.go` |
| `Sub` | `Sub`, `arith.go` (`Add(e.Neg())`) |
| `Quo`, `QuoExact`, `quoFint`, `quoBint` | `Quo` (`QuoExact` with scale 0), `quoFint`, `quoBint`, `arith.go` |
| `QuoRem`, `quoRemFint`, `quoRemBint` | same names, `arith.go` |
| `Cmp`, `cmpFint`, `cmpBint`, `Equal` | same names, `arith.go` |
| `fint`, `maxFint`, `pow10`, `add`, `mul`, `quo`, `quoRem`, `subAbs`, `lsh`, `fsa`, `isOdd`, `rshHalfEven`, `rshDown`, `prec`, `ntz`, `hasPrec` | `fint.go` |
| `bint`, `bpow10`, `bscale`, `mustParseBint`, `sign`, `cmp`, `string`, `setBint`, `setInt64`, `setFint`, `fint`, `add`, `inc`, `sub`, `subAbs`, `dbl`, `mul`, `pow`, `pow10`, `quo`, `quoRem`, `isOdd`, `lsh`, `fsa`, `rshHalfEven`, `prec`, `hasPrec`, `bpool`, `getBint`, `putBint` | `bint.go` |

Tests: `TestParse`, `TestNew`, `TestDecimal_String`, `TestDecimal_Prec`, `TestDecimal_Abs`,
`TestDecimal_Neg`, `TestDecimal_Add`, `TestDecimal_Sub`, `TestDecimal_Mul`, `TestDecimal_Quo`,
`TestDecimal_QuoRem`, `TestDecimal_Cmp` come from `decimal_test.go` upstream (now in
`decimal_test.go` and `arith_test.go`); the `TestFint_*` and `TestBint_*` tests come from
`integer_test.go` (now `fint_test.go`, `bint_test.go`). `TestBint_prec` generates the digit
table instead of listing it.

## Dropped

- `Sqrt`, `Exp`, `Expm1`, `Log`, `Log1p`, `Log2`, `Log10`, `Pow`, `PowInt`, `Inv`; the `bint`
  exp/log code and the tables `bfact`, `bexp`, `blog`, `bnlog10`.
- All serialization: JSON, Text, Binary, BSON, IEEE decimal128, `sql.Scanner`/`driver.Valuer`,
  `NullDecimal`; `NewFromFloat64`, `Float64`, `NewFromInt64`, `Int64`.
- `fmt.Formatter` (`Format`).
- `Sum`, `Mean`, `Prod`; `AddMul`, `SubMul`, `AddQuo`, `SubQuo`.
- All `*Exact` variants (`MulExact`, `AddExact`, `SubExact`, `QuoExact`, `ParseExact`).
- `SubAbs`, `Clamp`, `CmpTotal`, `CmpAbs`, `Less`, `Max`, `Min`, `CopySign`, `ULP`, `Quantize`,
  `SameScale`, `WithinOne`, `IsOne`, `MinScale` (method and constant), `Coef`, the `Zero()` and
  `One()` methods, `MustNew`, `NegOne`, `Two`, `Ten`, `Hundred`, `Thousand`, `E`, `Pi`.
- Rounding API `Round`, `Floor`, `Ceil`, `Trunc`, `Pad`, `Rescale`, `Trim` (only the unexported
  helpers `trunc`, `trim`, `minScale` remain, because `Quo` needs them); `fint.rshUp`.
- `unknownOverflowError`, `errInvalidOperation`, `bint.rshDown`, `bint.hlf`, `bint.setBytes`.
- Upstream `doc.go`, `doc_test.go` and all upstream fuzz tests.

## Differences from upstream

- `Parse` trims leading and trailing whitespace (`strings.TrimSpace`) before parsing;
  upstream rejects any whitespace.
- Exported error sentinels `ErrOverflow`, `ErrSyntax`, `ErrDivisionByZero`; every returned
  error satisfies `errors.Is` with one of them. A scale outside `0..19` in `New` is
  `ErrOverflow` with the message `decimal overflow: scale out of range`.
- The `scale` argument of `Add`, `Mul`, `Quo` and `Parse` does not exist; they behave as
  upstream with scale 0 (`AddExact(e, 0)`, ...).
- `Sub` error text names `+` with the negated operand (as upstream: it calls `AddExact(e.Neg(), 0)`).
- Unexported constant `scaleMin` replaces `MinScale`.
- Unit tests: `TestBint_prec` is generated; `rshUp` and `bint.rshDown` tests are gone with the code.
