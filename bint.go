// Derived from github.com/govalues/decimal v0.1.36 (MIT), see LICENSE.
package decimal

import (
	"fmt"
	"math/big"
	"sync"
)

// bint (Big INTeger) is a wrapper around big.Int.
type bint big.Int

// bpow10 is a cache of powers of 10, where bpow10[x] = 10^x.
var bpow10 = [...]*bint{
	mustParseBint("1"),
	mustParseBint("10"),
	mustParseBint("100"),
	mustParseBint("1000"),
	mustParseBint("10000"),
	mustParseBint("100000"),
	mustParseBint("1000000"),
	mustParseBint("10000000"),
	mustParseBint("100000000"),
	mustParseBint("1000000000"),
	mustParseBint("10000000000"),
	mustParseBint("100000000000"),
	mustParseBint("1000000000000"),
	mustParseBint("10000000000000"),
	mustParseBint("100000000000000"),
	mustParseBint("1000000000000000"),
	mustParseBint("10000000000000000"),
	mustParseBint("100000000000000000"),
	mustParseBint("1000000000000000000"),
	mustParseBint("10000000000000000000"),
	mustParseBint("100000000000000000000"),
	mustParseBint("1000000000000000000000"),
	mustParseBint("10000000000000000000000"),
	mustParseBint("100000000000000000000000"),
	mustParseBint("1000000000000000000000000"),
	mustParseBint("10000000000000000000000000"),
	mustParseBint("100000000000000000000000000"),
	mustParseBint("1000000000000000000000000000"),
	mustParseBint("10000000000000000000000000000"),
	mustParseBint("100000000000000000000000000000"),
	mustParseBint("1000000000000000000000000000000"),
	mustParseBint("10000000000000000000000000000000"),
	mustParseBint("100000000000000000000000000000000"),
	mustParseBint("1000000000000000000000000000000000"),
	mustParseBint("10000000000000000000000000000000000"),
	mustParseBint("100000000000000000000000000000000000"),
	mustParseBint("1000000000000000000000000000000000000"),
	mustParseBint("10000000000000000000000000000000000000"),
	mustParseBint("100000000000000000000000000000000000000"),
	mustParseBint("1000000000000000000000000000000000000000"),
	mustParseBint("10000000000000000000000000000000000000000"),
	mustParseBint("100000000000000000000000000000000000000000"),
	mustParseBint("1000000000000000000000000000000000000000000"),
	mustParseBint("10000000000000000000000000000000000000000000"),
	mustParseBint("100000000000000000000000000000000000000000000"),
	mustParseBint("1000000000000000000000000000000000000000000000"),
	mustParseBint("10000000000000000000000000000000000000000000000"),
	mustParseBint("100000000000000000000000000000000000000000000000"),
	mustParseBint("1000000000000000000000000000000000000000000000000"),
	mustParseBint("10000000000000000000000000000000000000000000000000"),
	mustParseBint("100000000000000000000000000000000000000000000000000"),
	mustParseBint("1000000000000000000000000000000000000000000000000000"),
	mustParseBint("10000000000000000000000000000000000000000000000000000"),
	mustParseBint("100000000000000000000000000000000000000000000000000000"),
	mustParseBint("1000000000000000000000000000000000000000000000000000000"),
	mustParseBint("10000000000000000000000000000000000000000000000000000000"),
	mustParseBint("100000000000000000000000000000000000000000000000000000000"),
	mustParseBint("1000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("10000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("100000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("1000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("10000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("100000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("1000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("10000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("100000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("1000000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("10000000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("100000000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("1000000000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("10000000000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("100000000000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("1000000000000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("10000000000000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("100000000000000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("1000000000000000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("10000000000000000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("100000000000000000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("1000000000000000000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("10000000000000000000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("100000000000000000000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("1000000000000000000000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("10000000000000000000000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("100000000000000000000000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("1000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("10000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("100000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("1000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("10000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("1000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("10000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("1000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("10000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("1000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("10000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("1000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),
}

// bscale is the number of fractional digits kept by quoBint before rounding.
const bscale = 41

// mustParseBint converts a string to *big.Int, panicking on error.
// Use only for package variable initialization and test code!
func mustParseBint(s string) *bint {
	z, ok := new(big.Int).SetString(s, 10)
	if !ok {
		panic(fmt.Errorf("mustParseBint(%q) failed: parsing error", s))
	}
	if z.Sign() < 0 {
		panic(fmt.Errorf("mustParseBint(%q) failed: negative number", s))
	}
	return (*bint)(z)
}

func (z *bint) sign() int {
	return (*big.Int)(z).Sign()
}

func (z *bint) cmp(x *bint) int {
	return (*big.Int)(z).Cmp((*big.Int)(x))
}

func (z *bint) string() string {
	return (*big.Int)(z).String()
}

func (z *bint) setBint(x *bint) {
	(*big.Int)(z).Set((*big.Int)(x))
}

func (z *bint) setInt64(x int64) {
	(*big.Int)(z).SetInt64(x)
}

func (z *bint) setFint(x fint) {
	(*big.Int)(z).SetUint64(uint64(x))
}

// fint converts *big.Int to uint64.
// If z cannot be represented as uint64, the result is undefined.
func (z *bint) fint() fint {
	f := (*big.Int)(z).Uint64()
	return fint(f)
}

// add calculates z = x + y.
func (z *bint) add(x, y *bint) {
	(*big.Int)(z).Add((*big.Int)(x), (*big.Int)(y))
}

// inc calcualtes z = x + 1.
func (z *bint) inc(x *bint) {
	y := bpow10[0]
	z.add(x, y)
}

// sub calculates z = x - y.
func (z *bint) sub(x, y *bint) {
	(*big.Int)(z).Sub((*big.Int)(x), (*big.Int)(y))
}

// subAbs calculates z = |x - y|.
func (z *bint) subAbs(x, y *bint) {
	switch x.cmp(y) {
	case 1:
		z.sub(x, y)
	default:
		z.sub(y, x)
	}
}

// dbl (Double) calculates z = x * 2.
func (z *bint) dbl(x *bint) {
	(*big.Int)(z).Lsh((*big.Int)(x), 1)
}

// mul calculates z = x * y.
func (z *bint) mul(x, y *bint) {
	// Copying x, y to prevent heap allocations.
	if z == x {
		b := getBint()
		defer putBint(b)
		b.setBint(x)
		x = b
	}
	if z == y {
		b := getBint()
		defer putBint(b)
		b.setBint(y)
		y = b
	}
	(*big.Int)(z).Mul((*big.Int)(x), (*big.Int)(y))
}

// pow calculates z = x^y.
// If y is negative, the result is unpredictable.
func (z *bint) pow(x, y *bint) {
	(*big.Int)(z).Exp((*big.Int)(x), (*big.Int)(y), nil)
}

// pow10 calculates z = 10^power.
// If power is negative, the result is unpredictable.
func (z *bint) pow10(power int) {
	x := getBint()
	defer putBint(x)
	x.setInt64(10)
	y := getBint()
	defer putBint(y)
	y.setInt64(int64(power))
	z.pow(x, y)
}

// quo calculates z = ⌊x / y⌋.
func (z *bint) quo(x, y *bint) {
	// Passing r to prevent heap allocations.
	r := getBint()
	defer putBint(r)
	z.quoRem(x, y, r)
}

// quoRem calculates z = ⌊x / y⌋, r = x - y * z.
func (z *bint) quoRem(x, y, r *bint) {
	(*big.Int)(z).QuoRem((*big.Int)(x), (*big.Int)(y), (*big.Int)(r))
}

func (z *bint) isOdd() bool {
	return (*big.Int)(z).Bit(0) != 0
}

// lsh (Left Shift) calculates z = x * 10^shift.
func (z *bint) lsh(x *bint, shift int) {
	var y *bint
	if shift < len(bpow10) {
		y = bpow10[shift]
	} else {
		y = getBint()
		defer putBint(y)
		y.pow10(shift)
	}
	z.mul(x, y)
}

// fsa (Fused Shift and Addition) calculates z = x * 10^shift + f.
func (z *bint) fsa(x *bint, shift int, f fint) {
	y := getBint()
	defer putBint(y)
	y.setFint(f)
	z.lsh(x, shift)
	z.add(z, y)
}

// rshHalfEven (Right Shift) calculates z = round(x / 10^shift) and
// rounds result using "half to even" rule.
func (z *bint) rshHalfEven(x *bint, shift int) {
	// Special cases
	switch {
	case x.sign() == 0:
		z.setFint(0)
		return
	case shift <= 0:
		z.setBint(x)
		return
	}
	// General case
	var y, r *bint
	r = getBint()
	defer putBint(r)
	if shift < len(bpow10) {
		y = bpow10[shift]
	} else {
		y = getBint()
		defer putBint(y)
		y.pow10(shift)
	}
	z.quoRem(x, y, r)
	r.dbl(r) // r = r * 2
	switch y.cmp(r) {
	case -1:
		z.inc(z) // z = z + 1
	case 0:
		// half-to-even
		if z.isOdd() {
			z.inc(z) // z = z + 1
		}
	}
}

// prec returns length of z in decimal digits.
// prec assumes that 0 has no digits.
// If z is negative, the result is unpredictable.
//
// z.prec() is significantly faster than len(z.string()),
// if z has less than len(bpow10) digits.
func (z *bint) prec() int {
	// Special case
	if z.cmp(bpow10[len(bpow10)-1]) > 0 {
		return len(z.string())
	}
	// General case
	left, right := 0, len(bpow10)
	for left < right {
		mid := (left + right) / 2
		if z.cmp(bpow10[mid]) < 0 {
			right = mid
		} else {
			left = mid + 1
		}
	}
	return left
}

// hasPrec checks if z has a given number of digits or more.
// hasPrec assumes that 0 has no digits.
// If z is negative, the result is unpredictable.
//
// z.hasPrec(p) is significantly faster than z.prec() >= p,
// if z has no more than len(bpow10) digits.
func (z *bint) hasPrec(prec int) bool {
	// Special cases
	switch {
	case prec < 1:
		return true
	case prec > len(bpow10):
		return len(z.string()) >= prec
	}
	// General case
	return z.cmp(bpow10[prec-1]) >= 0
}

// bpool is a cache of reusable *big.Int instances.
var bpool = sync.Pool{
	New: func() any {
		return (*bint)(new(big.Int))
	},
}

// getBint obtains a *big.Int from the pool.
func getBint() *bint {
	return bpool.Get().(*bint)
}

// putBint returns the *big.Int into the pool.
func putBint(b *bint) {
	bpool.Put(b)
}
