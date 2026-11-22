// Derived from github.com/govalues/decimal v0.1.36 (MIT), see LICENSE.
package decimal

import "fmt"

// Mul returns the (possibly rounded) product of decimals d and e.
//
// Mul returns an overflow error if the integer part of the result has
// more than [MaxPrec] digits.
func (d Decimal) Mul(e Decimal) (Decimal, error) {
	// General case
	f, err := d.mulFint(e, 0)
	if err != nil {
		f, err = d.mulBint(e, 0)
		if err != nil {
			return Decimal{}, fmt.Errorf("computing [%v * %v]: %w", d, e, err)
		}
	}
	return f, nil
}

// mulFint computes the product of two decimals using uint64 arithmetic.
func (d Decimal) mulFint(e Decimal, minScale int) (Decimal, error) {
	dcoef := d.coef
	dscale := d.Scale()
	dneg := d.IsNeg()

	ecoef := e.coef

	// Compute d = d * e
	dcoef, ok := dcoef.mul(ecoef)
	if !ok {
		return Decimal{}, ErrOverflow
	}
	dscale = dscale + e.Scale()
	dneg = dneg != e.IsNeg()

	return newFromFint(dneg, dcoef, dscale, minScale)
}

// mulBint computes the product of two decimals using *big.Int arithmetic.
func (d Decimal) mulBint(e Decimal, minScale int) (Decimal, error) {
	dcoef := getBint()
	defer putBint(dcoef)

	ecoef := getBint()
	defer putBint(ecoef)

	dcoef.setFint(d.coef)
	dscale := d.Scale()
	dneg := d.IsNeg()
	ecoef.setFint(e.coef)

	// Compute d = d * e
	dcoef.mul(dcoef, ecoef)
	dneg = dneg != e.IsNeg()
	dscale = dscale + e.Scale()

	return newFromBint(dneg, dcoef, dscale, minScale)
}

// Sub returns the (possibly rounded) difference between decimals d and e.
//
// Sub returns an error if the integer part of the result has more than [MaxPrec] digits.
func (d Decimal) Sub(e Decimal) (Decimal, error) {
	return d.Add(e.Neg())
}

// Add returns the (possibly rounded) sum of decimals d and e.
//
// Add returns an error if the integer part of the result has more than [MaxPrec] digits.
func (d Decimal) Add(e Decimal) (Decimal, error) {
	// General case
	f, err := d.addFint(e, 0)
	if err != nil {
		f, err = d.addBint(e, 0)
		if err != nil {
			return Decimal{}, fmt.Errorf("computing [%v + %v]: %w", d, e, err)
		}
	}

	return f, nil
}

// addFint computes the sum of two decimals using uint64 arithmetic.
func (d Decimal) addFint(e Decimal, minScale int) (Decimal, error) {
	dcoef := d.coef
	dscale := d.Scale()
	dneg := d.IsNeg()

	ecoef := e.coef

	// Alignment
	var ok bool
	switch {
	case dscale > e.Scale():
		ecoef, ok = ecoef.lsh(dscale - e.Scale())
		if !ok {
			return Decimal{}, ErrOverflow
		}
	case dscale < e.Scale():
		dcoef, ok = dcoef.lsh(e.Scale() - dscale)
		if !ok {
			return Decimal{}, ErrOverflow
		}
		dscale = e.Scale()
	}

	// Compute d = d + e
	if dneg == e.IsNeg() {
		dcoef, ok = dcoef.add(ecoef)
		if !ok {
			return Decimal{}, ErrOverflow
		}
	} else {
		if ecoef > dcoef {
			dneg = e.IsNeg()
		}
		dcoef = dcoef.subAbs(ecoef)
	}

	return newFromFint(dneg, dcoef, dscale, minScale)
}

// addBint computes the sum of two decimals using *big.Int arithmetic.
func (d Decimal) addBint(e Decimal, minScale int) (Decimal, error) {
	dcoef := getBint()
	defer putBint(dcoef)

	ecoef := getBint()
	defer putBint(ecoef)

	dcoef.setFint(d.coef)
	dscale := d.Scale()
	ecoef.setFint(e.coef)
	dneg := d.IsNeg()

	// Alignment
	switch {
	case dscale > e.Scale():
		ecoef.lsh(ecoef, dscale-e.Scale())
	case dscale < e.Scale():
		dcoef.lsh(dcoef, e.Scale()-dscale)
		dscale = e.Scale()
	}

	// Compute d = d + e
	if dneg == e.IsNeg() {
		dcoef.add(dcoef, ecoef)
	} else {
		if ecoef.cmp(dcoef) > 0 {
			dneg = e.IsNeg()
		}
		dcoef.subAbs(dcoef, ecoef)
	}

	return newFromBint(dneg, dcoef, dscale, minScale)
}

// Quo returns the (possibly rounded) quotient of decimals d and e.
//
// Quo returns an error if:
//   - the divisor is 0;
//   - the integer part of the result has more than [MaxPrec] digits.
func (d Decimal) Quo(e Decimal) (Decimal, error) {
	var scale int

	// Special case: zero divisor
	if e.IsZero() {
		return Decimal{}, fmt.Errorf("computing [%v / %v]: %w", d, e, ErrDivisionByZero)
	}

	// Special case: zero dividend
	if d.IsZero() {
		scale = max(scale, d.Scale()-e.Scale())
		return newSafe(false, 0, scale)
	}

	// General case
	f, err := d.quoFint(e, scale)
	if err != nil {
		f, err = d.quoBint(e, scale)
		if err != nil {
			return Decimal{}, fmt.Errorf("computing [%v / %v]: %w", d, e, err)
		}
	}

	// Preferred scale
	scale = max(scale, d.Scale()-e.Scale())
	f = f.trim(scale)

	return f, nil
}

// quoFint computes the quotient of two decimals using uint64 arithmetic.
func (d Decimal) quoFint(e Decimal, minScale int) (Decimal, error) {
	dcoef := d.coef
	dscale := d.Scale()
	dneg := d.IsNeg()

	ecoef := e.coef

	// Alignment
	var ok bool
	if shift := MaxPrec - dcoef.prec(); shift > 0 {
		dcoef, ok = dcoef.lsh(shift)
		if !ok {
			return Decimal{}, ErrOverflow // Should never happen
		}
		dscale = dscale + shift
	}
	if shift := ecoef.ntz(); shift > 0 {
		ecoef = ecoef.rshDown(shift)
		dscale = dscale + shift
	}

	// Compute d = d / e
	dcoef, ok = dcoef.quo(ecoef)
	if !ok {
		return Decimal{}, errInexactDivision
	}
	dscale = dscale - e.Scale()
	dneg = dneg != e.IsNeg()

	return newFromFint(dneg, dcoef, dscale, minScale)
}

// quoBint computes the quotient of two decimals using *big.Int arithmetic.
func (d Decimal) quoBint(e Decimal, minScale int) (Decimal, error) {
	dcoef := getBint()
	defer putBint(dcoef)

	ecoef := getBint()
	defer putBint(ecoef)

	dcoef.setFint(d.coef)
	dneg := d.IsNeg()
	ecoef.setFint(e.coef)

	// Alignment
	dcoef.lsh(dcoef, bscale+e.Scale()-d.Scale())

	// Compute d = ⌊d / e⌋
	dcoef.quo(dcoef, ecoef)
	dneg = dneg != e.IsNeg()

	return newFromBint(dneg, dcoef, bscale, minScale)
}

// QuoRem returns the quotient q and remainder r of decimals d and e
// such that d = e * q + r, where q is an integer and the sign of the
// reminder r is the same as the sign of the dividend d.
//
// QuoRem returns an error if:
//   - the divisor is 0;
//   - the integer part of the quotient has more than [MaxPrec] digits.
func (d Decimal) QuoRem(e Decimal) (q, r Decimal, err error) {
	// Special case: zero divisor
	if e.IsZero() {
		return Decimal{}, Decimal{}, fmt.Errorf("computing [%v div %v] and [%v mod %v]: %w", d, e, d, e, ErrDivisionByZero)
	}

	// General case
	q, r, err = d.quoRemFint(e)
	if err != nil {
		q, r, err = d.quoRemBint(e)
		if err != nil {
			return Decimal{}, Decimal{}, fmt.Errorf("computing [%v div %v] and [%v mod %v]: %w", d, e, d, e, err)
		}
	}

	return q, r, nil
}

// quoRemFint computes the quotient and remainder of two decimals using uint64 arithmetic.
func (d Decimal) quoRemFint(e Decimal) (q, r Decimal, err error) {
	dcoef := d.coef
	ecoef := e.coef
	rscale := d.Scale()

	// Alignment
	var ok bool
	switch {
	case d.Scale() > e.Scale():
		ecoef, ok = ecoef.lsh(d.Scale() - e.Scale())
		if !ok {
			return Decimal{}, Decimal{}, ErrOverflow
		}
	case d.Scale() < e.Scale():
		dcoef, ok = dcoef.lsh(e.Scale() - d.Scale())
		if !ok {
			return Decimal{}, Decimal{}, ErrOverflow
		}
		rscale = e.Scale()
	}

	// Compute q = ⌊d / e⌋, r = d - e * q
	qcoef, rcoef, ok := dcoef.quoRem(ecoef)
	if !ok {
		return Decimal{}, Decimal{}, ErrDivisionByZero // Should never happen
	}
	qsign := d.IsNeg() != e.IsNeg()
	rsign := d.IsNeg()

	q, err = newFromFint(qsign, qcoef, 0, 0)
	if err != nil {
		return Decimal{}, Decimal{}, err
	}
	r, err = newFromFint(rsign, rcoef, rscale, rscale)
	if err != nil {
		return Decimal{}, Decimal{}, err
	}
	return q, r, nil
}

// quoRemBint computes the quotient and remainder of two decimals using *big.Int arithmetic.
func (d Decimal) quoRemBint(e Decimal) (q, r Decimal, err error) {
	dcoef := getBint()
	defer putBint(dcoef)

	ecoef := getBint()
	defer putBint(ecoef)

	qcoef := getBint()
	defer putBint(qcoef)

	rcoef := getBint()
	defer putBint(rcoef)

	dcoef.setFint(d.coef)
	ecoef.setFint(e.coef)
	rscale := d.Scale()

	// Alignment
	switch {
	case d.Scale() > e.Scale():
		ecoef.lsh(ecoef, d.Scale()-e.Scale())
	case d.Scale() < e.Scale():
		dcoef.lsh(dcoef, e.Scale()-d.Scale())
		rscale = e.Scale()
	}

	// Compute q = ⌊d / e⌋, r = d - e * q
	qcoef.quoRem(dcoef, ecoef, rcoef)
	qsign := d.IsNeg() != e.IsNeg()
	rsign := d.IsNeg()

	q, err = newFromBint(qsign, qcoef, 0, 0)
	if err != nil {
		return Decimal{}, Decimal{}, err
	}
	r, err = newFromBint(rsign, rcoef, rscale, rscale)
	if err != nil {
		return Decimal{}, Decimal{}, err
	}
	return q, r, nil
}

// Equal compares decimals and returns:
//
//	 true if d = e
//	false otherwise
//
// See also method [Decimal.Cmp].
func (d Decimal) Equal(e Decimal) bool {
	return d.Cmp(e) == 0
}

// Cmp compares decimals and returns:
//
//	-1 if d < e
//	 0 if d = e
//	+1 if d > e
func (d Decimal) Cmp(e Decimal) int {
	// Special case: different signs
	switch {
	case d.Sign() > e.Sign():
		return 1
	case d.Sign() < e.Sign():
		return -1
	}

	// General case
	r, err := d.cmpFint(e)
	if err != nil {
		r = d.cmpBint(e)
	}
	return r
}

// cmpFint compares decimals using uint64 arithmetic.
func (d Decimal) cmpFint(e Decimal) (int, error) {
	dcoef := d.coef
	ecoef := e.coef

	// Alignment
	var ok bool
	switch {
	case d.Scale() > e.Scale():
		ecoef, ok = ecoef.lsh(d.Scale() - e.Scale())
		if !ok {
			return 0, ErrOverflow
		}
	case d.Scale() < e.Scale():
		dcoef, ok = dcoef.lsh(e.Scale() - d.Scale())
		if !ok {
			return 0, ErrOverflow
		}
	}

	// Comparison
	switch {
	case dcoef > ecoef:
		return d.Sign(), nil
	case ecoef > dcoef:
		return -e.Sign(), nil
	}
	return 0, nil
}

// cmpBint compares decimals using *big.Int arithmetic.
func (d Decimal) cmpBint(e Decimal) int {
	dcoef := getBint()
	defer putBint(dcoef)

	ecoef := getBint()
	defer putBint(ecoef)

	dcoef.setFint(d.coef)
	ecoef.setFint(e.coef)

	// Alignment
	switch {
	case d.Scale() > e.Scale():
		ecoef.lsh(ecoef, d.Scale()-e.Scale())
	case d.Scale() < e.Scale():
		dcoef.lsh(dcoef, e.Scale()-d.Scale())
	}

	// Comparison
	switch dcoef.cmp(ecoef) {
	case 1:
		return d.Sign()
	case -1:
		return -e.Sign()
	}
	return 0
}
