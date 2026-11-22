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
- Serialization except JSON and Text (see "Encoding" below): Binary, BSON, IEEE decimal128, `sql.Scanner`/`driver.Valuer`,
  `NullDecimal`; upstream's `NewFromInt64` and `Int64` (`NewFromFloat64`, `Float64` and a
  different `Int64` exist here, see "New code").
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
- Rounding is a `RoundingMode` API (`Round(scale, mode)`, `RoundStep`) instead of the upstream
  methods `Round`, `Floor`, `Ceil`, `Trunc`, `Pad`, `Rescale`, `Trim`. `Round` never pads;
  `Rescale` and `Trim` are new and differ from upstream's `Rescale`/`Trim`.

## Encoding (v0.2.0, adapted from upstream)

`encoding.go` carries the attribution header: `UnmarshalJSON`, `UnmarshalText`, `MarshalText`
and `AppendText` follow upstream's methods of the same names. Differences: `MarshalJSON` writes
a bare JSON number, upstream a quoted string; `UnmarshalJSON` rejects `null` (error matching
`ErrSyntax`), upstream ignores it; strings with escapes are rejected; surrounding whitespace is
trimmed as in `Parse`; on error the receiver is left unchanged; `AppendText` has the standard
`encoding.TextAppender` signature `([]byte, error)`. Still dropped: Binary, BSON, IEEE 754
decimal128, `sql.Scanner`, `driver.Valuer`, `NullDecimal`.

## New code (not derived from upstream)

These files contain no upstream code and carry no attribution header:

| File | Contents |
| --- | --- |
| `round.go` | `RoundingMode`, `Round`, `Rescale`, `Trim`, `RoundStep`, `IsMultipleOf`, `ErrInvalidStep`, `ErrInvalidMode`, `ErrInexact` |
| `convert.go` | `NewFromFloat64`, `Float64`, `Int64` |
| `quo.go` | `QuoRound` (v0.2.0) |
| `minmax.go` | `Min`, `Max` (v0.2.0) |
| `quo_test.go`, `quo_fuzz_test.go`, `encoding_test.go` | tests and `FuzzQuoRound` (v0.2.0) |
| `fuzz_test.go`, `round_test.go`, `convert_test.go`, `bench_test.go` | tests, fuzzers, benchmarks, allocation checks |

`round.go` uses the ported `fint`/`bint` helpers (`quoRem`, `lsh`, `mul`, `dbl`, `inc`, ...) and the
ported `trim`; `Trim` is a wrapper over it. `NewFromFloat64` uses the ported `parse`, and
`Float64` is upstream's `String` + `strconv.ParseFloat` approach (upstream `Float64`, same
semantics: nearest float, `ok` false only on a conversion error). `NewFromFloat64` differs from
upstream (which used `strconv.AppendFloat(..., 'f', -1, 64)`): it uses `'g'`, so no 300-digit
strings are built for large or tiny floats.

## Benchmarks

`go test -run '^$' -bench . -benchmem .` on Apple M1 Pro, darwin/arm64, Go 1.27:

| Benchmark | ns/op | B/op | allocs/op |
| --- | ---: | ---: | ---: |
| Parse `65000.1` | 26.03 | 0 | 0 |
| Parse `0.012` | 21.49 | 0 | 0 |
| String | 18.77 | 8 | 1 |
| Add | 6.02 | 0 | 0 |
| Mul | 5.58 | 0 | 0 |
| Quo exact (`65000.17 / 0.01`) | 33.53 | 0 | 0 |
| Quo inexact (`65000.17 / 0.012`) | 315.4 | 40 | 2 |
| QuoRound `65000.17 / 3`, scale 8, HalfEven | 13.94 | 0 | 0 |
| QuoRound `1 / 3`, scale 19, Floor | 10.14 | 0 | 0 |
| QuoRound `0.9999999999999999999 / 7`, scale 19, HalfEven (`*big.Int`) | 360.0 | 40 | 2 |
| MarshalJSON | 9.76 | 0 | 0 |
| UnmarshalJSON `65000.17` | 34.71 | 0 | 0 |
| Cmp | 6.75 | 0 | 0 |
| RoundStep `65000.17` step `0.1` HalfUp | 7.34 | 0 | 0 |
| RoundStep `0.01234` step `0.001` Floor | 7.87 | 0 | 0 |

An inexact `Quo` takes the `*big.Int` path (upstream behaviour) and allocates. `QuoRound`
allocates only when the aligned operands do not fit uint64. The benchmark discards the
`MarshalJSON` result, which is why it shows 0 allocs; real calls allocate the returned slice.
