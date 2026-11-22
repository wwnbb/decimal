// Derived from github.com/govalues/decimal v0.1.36 (MIT), see LICENSE.
package decimal

// fint (Fast INTeger) is a wrapper around uint64.
type fint uint64

// maxFint is a maximum value of fint.
const maxFint = 9_999_999_999_999_999_999

// pow10 is a cache of powers of 10, where pow10[x] = 10^x.
var pow10 = [...]fint{
	1,                          // 10^0
	10,                         // 10^1
	100,                        // 10^2
	1_000,                      // 10^3
	10_000,                     // 10^4
	100_000,                    // 10^5
	1_000_000,                  // 10^6
	10_000_000,                 // 10^7
	100_000_000,                // 10^8
	1_000_000_000,              // 10^9
	10_000_000_000,             // 10^10
	100_000_000_000,            // 10^11
	1_000_000_000_000,          // 10^12
	10_000_000_000_000,         // 10^13
	100_000_000_000_000,        // 10^14
	1_000_000_000_000_000,      // 10^15
	10_000_000_000_000_000,     // 10^16
	100_000_000_000_000_000,    // 10^17
	1_000_000_000_000_000_000,  // 10^18
	10_000_000_000_000_000_000, // 10^19
}

// add calculates x + y and checks overflow.
func (x fint) add(y fint) (z fint, ok bool) {
	if maxFint-x < y {
		return 0, false
	}
	z = x + y
	return z, true
}

// mul calculates x * y and checks overflow.
func (x fint) mul(y fint) (z fint, ok bool) {
	if y == 0 {
		return 0, true
	}
	z = x * y
	if z/y != x {
		return 0, false
	}
	if z > maxFint {
		return 0, false
	}
	return z, true
}

// quo calculates x / y and checks division by zero and inexact division.
func (x fint) quo(y fint) (z fint, ok bool) {
	if y == 0 {
		return 0, false
	}
	z = x / y
	if z*y != x {
		return 0, false
	}
	return z, true
}

// quoRem calculates q = ⌊x / y⌋, r = x - y * q and checks division by zero.
func (x fint) quoRem(y fint) (q, r fint, ok bool) {
	if y == 0 {
		return 0, 0, false
	}
	q = x / y
	r = x - q*y
	return q, r, true
}

// subAbs calculates |x - y|.
func (x fint) subAbs(y fint) fint {
	if x > y {
		return x - y
	}
	return y - x
}

// lsh (Left Shift) calculates x * 10^shift and checks overflow.
func (x fint) lsh(shift int) (z fint, ok bool) {
	// Special cases
	switch {
	case shift <= 0:
		return x, true
	case shift == 1 && x < maxFint/10: // to speed up common case
		return x * 10, true
	case shift >= len(pow10):
		return 0, false
	}
	// General case
	y := pow10[shift]
	return x.mul(y)
}

// fsa (Fused Shift and Addition) calculates x * 10^shift + b and checks overflow.
func (x fint) fsa(shift int, b byte) (z fint, ok bool) {
	z, ok = x.lsh(shift)
	if !ok {
		return 0, false
	}
	z, ok = z.add(fint(b))
	if !ok {
		return 0, false
	}
	return z, true
}

func (x fint) isOdd() bool {
	return x&1 != 0
}

// rshHalfEven (Right Shift) calculates round(x / 10^shift) and rounds result
// using "half to even" rule.
func (x fint) rshHalfEven(shift int) fint {
	// Special cases
	switch {
	case x == 0:
		return 0
	case shift <= 0:
		return x
	case shift >= len(pow10):
		return 0
	}
	// General case
	y := pow10[shift]
	z := x / y
	r := x - z*y                        // r = x % y
	y = y >> 1                          // y = y / 2, which is safe as y is a multiple of 10
	if y < r || (y == r && z.isOdd()) { // half-to-even
		z++
	}
	return z
}

// rshDown (Right Shift) calculates ⌊x / 10^shift⌋ and rounds result towards zero.
func (x fint) rshDown(shift int) fint {
	// Special cases
	switch {
	case x == 0:
		return 0
	case shift <= 0:
		return x
	case shift >= len(pow10):
		return 0
	}
	// General case
	y := pow10[shift]
	return x / y
}

// prec returns length of x in decimal digits.
// prec assumes that 0 has no digits.
func (x fint) prec() int {
	left, right := 0, len(pow10)
	for left < right {
		mid := (left + right) / 2
		if x < pow10[mid] {
			right = mid
		} else {
			left = mid + 1
		}
	}
	return left
}

// ntz returns number of trailing zeros in x.
// ntz assumes that 0 has no trailing zeros.
func (x fint) ntz() int {
	left, right := 1, x.prec()
	for left < right {
		mid := (left + right) / 2
		if x%pow10[mid] == 0 {
			left = mid + 1
		} else {
			right = mid
		}
	}
	return left - 1
}

// hasPrec returns true if x has given number of digits or more.
// hasPrec assumes that 0 has no digits.
//
// x.hasPrec(p) is significantly faster than x.prec() >= p.
func (x fint) hasPrec(prec int) bool {
	// Special cases
	switch {
	case prec < 1:
		return true
	case prec > len(pow10):
		return false
	}
	// General case
	return x >= pow10[prec-1]
}
