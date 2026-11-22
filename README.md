# decimal

Exact decimal arithmetic for prices, quantities and money: parsing and
printing decimal text, the four operations, comparison, explicit rounding
modes and rounding to an instrument step. No dependencies beyond the
standard library, no heap allocations on the fast path, no panics.

The representation and the arithmetic core are derived from
[govalues/decimal](https://github.com/govalues/decimal) v0.1.36 (MIT, see
LICENSE); NOTICE.md lists what was taken and what differs.

A value has up to 19 significant digits and up to 19 digits after the
decimal point. An exact result that needs more digits is rounded half to
even; an integer part that does not fit is an error.
