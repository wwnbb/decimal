package decimal

// Min returns the smaller of a and b by value. If they are equal, it returns a.
func Min(a, b Decimal) Decimal {
	if b.Cmp(a) < 0 {
		return b
	}
	return a
}

// Max returns the larger of a and b by value. If they are equal, it returns a.
func Max(a, b Decimal) Decimal {
	if b.Cmp(a) > 0 {
		return b
	}
	return a
}
