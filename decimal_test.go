// Derived from github.com/govalues/decimal v0.1.36 (MIT), see LICENSE.
package decimal

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"testing"
	"unsafe"
)

func TestDecimal_ZeroValue(t *testing.T) {
	got := Decimal{}
	want := mustNew(0, 0)
	if got != want || got != Zero {
		t.Errorf("Decimal{} = %q, want %q", got, want)
	}
}

func TestDecimal_Size(t *testing.T) {
	d := Decimal{}
	got := unsafe.Sizeof(d)
	want := uintptr(16)
	if got != want {
		t.Errorf("unsafe.Sizeof(%q) = %v, want %v", d, got, want)
	}
}

func TestDecimal_Interfaces(t *testing.T) {
	var d any = Decimal{}
	if _, ok := d.(fmt.Stringer); !ok {
		t.Errorf("%T does not implement fmt.Stringer", d)
	}
}

func TestConstants(t *testing.T) {
	if got := One.String(); got != "1" {
		t.Errorf("One = %q, want 1", got)
	}
	if One != mustNew(1, 0) {
		t.Errorf("One = %+v, want New(1, 0)", One)
	}
	if Zero.String() != "0" || !Zero.IsZero() {
		t.Errorf("Zero = %q", Zero)
	}
	if MaxPrec != 19 || MaxScale != 19 {
		t.Errorf("MaxPrec, MaxScale = %v, %v, want 19, 19", MaxPrec, MaxScale)
	}
}

// mustNew is a test helper, the library has no MustNew.
func mustNew(value int64, scale int) Decimal {
	d, err := New(value, scale)
	if err != nil {
		panic(fmt.Sprintf("New(%v, %v) failed: %v", value, scale, err))
	}
	return d
}

func TestNew(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			value int64
			scale int
			want  string
		}{
			{math.MinInt64, 0, "-9223372036854775808"},
			{math.MinInt64, 1, "-922337203685477580.8"},
			{math.MinInt64, 2, "-92233720368547758.08"},
			{math.MinInt64, 19, "-0.9223372036854775808"},
			{0, 0, "0"},
			{0, 1, "0.0"},
			{0, 2, "0.00"},
			{0, 3, "0.000"},
			{0, 19, "0.0000000000000000000"},
			{1, 0, "1"},
			{1, 1, "0.1"},
			{1, 2, "0.01"},
			{1, 19, "0.0000000000000000001"},
			{math.MaxInt64, 0, "9223372036854775807"},
			{math.MaxInt64, 1, "922337203685477580.7"},
			{math.MaxInt64, 2, "92233720368547758.07"},
			{math.MaxInt64, 19, "0.9223372036854775807"},
		}
		for _, tt := range tests {
			got, err := New(tt.value, tt.scale)
			if err != nil {
				t.Errorf("New(%v, %v) failed: %v", tt.value, tt.scale, err)
				continue
			}
			want := MustParse(tt.want)
			if got != want {
				t.Errorf("New(%v, %v) = %q, want %q", tt.value, tt.scale, got, want)
			}
		}
	})

	t.Run("error", func(t *testing.T) {
		tests := map[string]struct {
			value int64
			scale int
		}{
			"scale range 1": {math.MinInt64, -1},
			"scale range 2": {math.MaxInt64, -1},
			"scale range 3": {0, -1},
			"scale range 4": {1, -1},
			"scale range 5": {math.MinInt64, 20},
			"scale range 6": {math.MinInt64, 39},
			"scale range 7": {math.MaxInt64, 20},
			"scale range 8": {math.MaxInt64, 39},
		}
		for name, tt := range tests {
			_, err := New(tt.value, tt.scale)
			if err == nil {
				t.Errorf("%v: New(%v, %v) did not fail", name, tt.value, tt.scale)
				continue
			}
			if !errors.Is(err, ErrOverflow) {
				t.Errorf("%v: New(%v, %v) error %q is not ErrOverflow", name, tt.value, tt.scale, err)
			}
		}
	})
}

func TestParse(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			s         string
			wantNeg   bool
			wantCoef  uint64
			wantScale int
		}{
			{"-9999999999999999999.0", true, 9999999999999999999, 0},
			{"-9999999999999999999", true, 9999999999999999999, 0},
			{"-999999999999999999.9", true, 9999999999999999999, 1},
			{"-99999999999999999.99", true, 9999999999999999999, 2},
			{"-1000000000000000000.0", true, 1000000000000000000, 0},
			{"-0.9999999999999999999", true, 9999999999999999999, 19},
			{"-00000000000000000000000000000000000001", true, 1, 0},
			{"-1", true, 1, 0},
			{"-1.", true, 1, 0},
			{"-.1", true, 1, 1},
			{"-0.1", true, 1, 1},
			{"-0.01", true, 1, 2},
			{"-0.0000000000000000001", true, 1, 19},
			{"-00000000000000000000000000000000000000", false, 0, 0},
			{"+00000000000000000000000000000000000000", false, 0, 0},
			{"0", false, 0, 0},
			{"0.", false, 0, 0},
			{".0", false, 0, 1},
			{"0.0", false, 0, 1},
			{"0.00", false, 0, 2},
			{"0.000000000000000000", false, 0, 18},
			{"0.0000000000000000000", false, 0, 19},
			{"0.00000000000000000000", false, 0, 19},
			{"00000000000000000000000000000000000001", false, 1, 0},
			{"1", false, 1, 0},
			{"1.", false, 1, 0},
			{".1", false, 1, 1},
			{"0.1", false, 1, 1},
			{"0.01", false, 1, 2},
			{"0.0000000000000000001", false, 1, 19},
			{"1000000000000000000.0", false, 1000000000000000000, 0},
			{"9999999999999999999.0", false, 9999999999999999999, 0},
			{"9999999999999999999", false, 9999999999999999999, 0},
			{"999999999999999999.9", false, 9999999999999999999, 1},
			{"99999999999999999.99", false, 9999999999999999999, 2},
			{"0.9999999999999999999", false, 9999999999999999999, 19},

			// Rounding
			{"-999999999999999999.99", true, 1000000000000000000, 0},
			{"0.00000000000000000000000000000000000000", false, 0, 19},
			{"0.00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000", false, 0, 19},
			{"-0.00000000000000000000000000000000000001", false, 0, 19},
			{"0.00000000000000000000000000000000000001", false, 0, 19},
			{"0.123456789012345678901234567890", false, 1234567890123456789, 19},
			{"0.12345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678", false, 1234567890123456789, 19},

			// Exponential notation
			{"0e9", false, 0, 0},
			{"0e-9", false, 0, 9},
			{"1.23e-12", false, 123, 14},
			{"1.23e-5", false, 123, 7},
			{"1.23e-4", false, 123, 6},
			{"1.23e-3", false, 123, 5},
			{"1.23e-2", false, 123, 4},
			{"1.23e-1", false, 123, 3},
			{"1.23e+0", false, 123, 2},
			{"1.23e+1", false, 123, 1},
			{"1.23e+2", false, 123, 0},
			{"1.23e+3", false, 1230, 0},
			{"1.23e+4", false, 12300, 0},
			{"1.23e+5", false, 123000, 0},
			{"1.23e+12", false, 1230000000000, 0},
			{"0.0e-38", false, 0, 19},
			{"0e-38", false, 0, 19},
			{"1e-2", false, 1, 2},
			{"1e-1", false, 1, 1},
			{"1e0", false, 1, 0},
			{"1e+1", false, 10, 0},
			{"1e+2", false, 100, 0},
			{"0.0000000000000000001e-19", false, 0, 19},
			{"0.0000000000000000001e19", false, 1, 0},
			{"1000000000000000000e-19", false, 1000000000000000000, 19},
			{"1000000000000000000e-38", false, 0, 19},
			{"10000000000000000000e-38", false, 1, 19},
			{"100000000000000000000e-38", false, 10, 19},
			{"10000000000000000000000000000000000000e-38", false, 1000000000000000000, 19},
			{"1e+18", false, 1000000000000000000, 0},
			{"0.0000000001e10", false, 1, 0},
			{"10000000000e-10", false, 10000000000, 10},
			{"4E9", false, 4000000000, 0},
			{"0.73e-7", false, 73, 9},

			// Surrounding whitespace is trimmed (differs from upstream)
			{" 1", false, 1, 0},
			{" +1", false, 1, 0},
			{" -1", true, 1, 0},
			{"1 ", false, 1, 0},
			{"+1 ", false, 1, 0},
			{"-1 ", true, 1, 0},
			{" 1 ", false, 1, 0},
			{"\t1.50\n", false, 150, 2},
			{"\r\n -0.5e1 \t", true, 5, 0},

			// Signs, exponents, zeros
			{"+1", false, 1, 0},
			{"1e-5", false, 1, 5},
			{"1E3", false, 1000, 0},
			{"-0.0", false, 0, 1},
			{"-0", false, 0, 0},
			{"-0e5", false, 0, 0},
		}
		for _, tt := range tests {
			got, err := Parse(tt.s)
			if err != nil {
				t.Errorf("Parse(%q) failed: %v", tt.s, err)
				continue
			}
			if got.IsNeg() != tt.wantNeg {
				t.Errorf("Parse(%q).IsNeg() = %v, want %v", tt.s, got.IsNeg(), tt.wantNeg)
				continue
			}
			if uint64(got.coef) != tt.wantCoef {
				t.Errorf("Parse(%q).coef = %v, want %v", tt.s, got.coef, tt.wantCoef)
				continue
			}
			if got.Scale() != tt.wantScale {
				t.Errorf("Parse(%q).Scale() = %v, want %v", tt.s, got.Scale(), tt.wantScale)
				continue
			}
		}
	})

	t.Run("string", func(t *testing.T) {
		tests := []struct{ s, want string }{
			{"1e-5", "0.00001"},
			{"1E3", "1000"},
			{"-0.0", "0.0"},
			{"+1", "1"},
			{" 1.50 ", "1.50"},
			{"1.5e1", "15"},
			{"15e-1", "1.5"},
		}
		for _, tt := range tests {
			got := MustParse(tt.s).String()
			if got != tt.want {
				t.Errorf("Parse(%q).String() = %q, want %q", tt.s, got, tt.want)
			}
		}
	})

	t.Run("negative zero is zero", func(t *testing.T) {
		d := MustParse("-0.0")
		if !d.IsZero() || d.IsNeg() || d.Sign() != 0 || d.IsPos() {
			t.Errorf("Parse(-0.0) = %+v, want non-negative zero", d)
		}
	})

	t.Run("error", func(t *testing.T) {
		tests := map[string]struct {
			s    string
			want error
		}{
			"missing digits 1":  {"", ErrSyntax},
			"missing digits 2":  {"+", ErrSyntax},
			"missing digits 3":  {"-", ErrSyntax},
			"missing digits 4":  {".", ErrSyntax},
			"missing digits 5":  {"..", ErrSyntax},
			"missing digits 6":  {".e", ErrSyntax},
			"missing digits 7":  {"e1", ErrSyntax},
			"missing digits 8":  {"+e", ErrSyntax},
			"missing digits 9":  {"-e", ErrSyntax},
			"missing digits 10": {"e+", ErrSyntax},
			"missing digits 11": {"e-", ErrSyntax},
			"missing digits 12": {"e.0", ErrSyntax},
			"missing digits 13": {"e+1", ErrSyntax},
			"missing digits 14": {"e-1", ErrSyntax},
			"blank 1":           {" ", ErrSyntax},
			"blank 2":           {" \t\n", ErrSyntax},
			"invalid char 1":    {"a", ErrSyntax},
			"invalid char 2":    {"1a", ErrSyntax},
			"invalid char 3":    {"1.a", ErrSyntax},
			"invalid char 4":    {" + 1", ErrSyntax},
			"invalid char 5":    {" - 1", ErrSyntax},
			"invalid char 6":    {"1,1", ErrSyntax},
			"invalid char 7":    {"1 1", ErrSyntax},
			"invalid char 8":    {"1/3", ErrSyntax},
			"invalid char 9":    {"0x10", ErrSyntax},
			"invalid char 10":   {"1_000", ErrSyntax},
			"missing exp 1":     {"0.e", ErrSyntax},
			"missing exp 2":     {"1e", ErrSyntax},
			"missing exp 3":     {"1ee", ErrSyntax},
			"exp range 1":       {"1e-331", ErrSyntax},
			"exp range 2":       {"1e331", ErrSyntax},
			"exp range 3":       {"1e999", ErrSyntax},
			"double sign 1":     {"++1", ErrSyntax},
			"double sign 2":     {"--1", ErrSyntax},
			"double sign 3":     {"+-1", ErrSyntax},
			"double sign 4":     {"-+1", ErrSyntax},
			"double sign 5":     {"-1.-1", ErrSyntax},
			"double sign 6":     {"1.1-", ErrSyntax},
			"double sign 7":     {"1e--1", ErrSyntax},
			"double sign 8":     {"1e-+1", ErrSyntax},
			"double sign 9":     {"1e+-1", ErrSyntax},
			"double sign 10":    {"1e++1", ErrSyntax},
			"double sign 11":    {"1e-1-", ErrSyntax},
			"double sign 12":    {"-1-", ErrSyntax},
			"double dot 1":      {"1.1.1", ErrSyntax},
			"double dot 2":      {"..1", ErrSyntax},
			"double dot 3":      {"1..1", ErrSyntax},
			"double dot 4":      {".1.1", ErrSyntax},
			"double dot 5":      {"1.1.", ErrSyntax},
			"double dot 6":      {".1.", ErrSyntax},
			"special value 1":   {"Inf", ErrSyntax},
			"special value 2":   {"-infinity", ErrSyntax},
			"special value 3":   {"NaN", ErrSyntax},
			"special value 4":   {"+Inf", ErrSyntax},
			"special value 5":   {"nan", ErrSyntax},
			"overflow 1":        {"-10000000000000000000", ErrOverflow},
			"overflow 2":        {"-99999999999999999990", ErrOverflow},
			"overflow 3":        {"10000000000000000000", ErrOverflow},
			"overflow 4":        {"99999999999999999990", ErrOverflow},
			"overflow 5":        {"123456789012345678901234567890123456789", ErrOverflow},
			"overflow 6":        {"1e19", ErrOverflow},
			"overflow 7":        {"1e330", ErrOverflow},
			"many digits":       {"0." + strings.Repeat("0", 400), ErrSyntax},
		}
		for name, tt := range tests {
			t.Run(name, func(t *testing.T) {
				_, err := Parse(tt.s)
				if err == nil {
					t.Errorf("Parse(%q) did not fail", tt.s)
					return
				}
				if !errors.Is(err, tt.want) {
					t.Errorf("Parse(%q) error %q does not match %q", tt.s, err, tt.want)
				}
			})
		}
	})
}

func TestMustParse(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		if got := MustParse("1.50").String(); got != "1.50" {
			t.Errorf("MustParse(1.50) = %q", got)
		}
	})
	t.Run("error", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("MustParse(\".\") did not panic")
			}
		}()
		MustParse(".")
	})
}

func TestDecimal_String(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			neg   bool
			coef  fint
			scale int
			want  string
		}{
			{true, maxCoef, 0, "-9999999999999999999"},
			{true, maxCoef, 1, "-999999999999999999.9"},
			{true, maxCoef, 2, "-99999999999999999.99"},
			{true, maxCoef, 3, "-9999999999999999.999"},
			{true, maxCoef, 19, "-0.9999999999999999999"},
			{true, 1, 0, "-1"},
			{true, 1, 1, "-0.1"},
			{true, 1, 2, "-0.01"},
			{true, 1, 19, "-0.0000000000000000001"},
			{false, 0, 0, "0"},
			{false, 0, 1, "0.0"},
			{false, 0, 2, "0.00"},
			{false, 0, 19, "0.0000000000000000000"},
			{false, 1, 19, "0.0000000000000000001"},
			{false, 1, 2, "0.01"},
			{false, 1, 1, "0.1"},
			{false, 1, 0, "1"},
			{false, maxCoef, 19, "0.9999999999999999999"},
			{false, maxCoef, 3, "9999999999999999.999"},
			{false, maxCoef, 2, "99999999999999999.99"},
			{false, maxCoef, 1, "999999999999999999.9"},
			{false, maxCoef, 0, "9999999999999999999"},

			// Exported constants
			{Zero.neg, Zero.coef, Zero.Scale(), "0"},
			{One.neg, One.coef, One.Scale(), "1"},
		}
		for _, tt := range tests {
			d, err := newSafe(tt.neg, tt.coef, tt.scale)
			if err != nil {
				t.Errorf("newDecimal(%v, %v, %v) failed: %v", tt.neg, tt.coef, tt.scale, err)
				continue
			}
			got := d.String()
			if got != tt.want {
				t.Errorf("newDecimal(%v, %v, %v).String() = %q, want %q", tt.neg, tt.coef, tt.scale, got, tt.want)
			}
		}
	})

	t.Run("round trip", func(t *testing.T) {
		tests := []string{
			"0", "0.0", "0.00", "1", "-1", "1.50", "-1.50", "0.1", "-0.1", "0.00001",
			"1000", "100.001", "9999999999999999999", "-9999999999999999999",
			"0.9999999999999999999", "-0.0000000000000000001", "0.0000000000000000000",
			"123456789.123456789", "-92233720368547758.08",
		}
		for _, s := range tests {
			if got := MustParse(s).String(); got != s {
				t.Errorf("String(MustParse(%q)) = %q", s, got)
			}
		}
	})
}

func TestDecimal_Prec(t *testing.T) {
	tests := []struct {
		d    string
		want int
	}{
		{"0000", 0},
		{"000", 0},
		{"00", 0},
		{"0", 0},
		{"0.000", 0},
		{"0.00", 0},
		{"0.0", 0},
		{"0", 0},
		{"0.0000000000000000001", 1},
		{"0.000000000000000001", 1},
		{"0.00000000000000001", 1},
		{"0.0000000000000001", 1},
		{"0.000000000000001", 1},
		{"0.00000000000001", 1},
		{"0.0000000000001", 1},
		{"0.000000000001", 1},
		{"0.00000000001", 1},
		{"0.0000000001", 1},
		{"0.000000001", 1},
		{"0.00000001", 1},
		{"0.0000001", 1},
		{"0.000001", 1},
		{"0.00001", 1},
		{"0.0001", 1},
		{"0.001", 1},
		{"0.01", 1},
		{"0.1", 1},
		{"1", 1},
		{"0.1000000000000000000", 19},
		{"0.100000000000000000", 18},
		{"0.10000000000000000", 17},
		{"0.1000000000000000", 16},
		{"0.100000000000000", 15},
		{"0.10000000000000", 14},
		{"0.1000000000000", 13},
		{"0.100000000000", 12},
		{"0.10000000000", 11},
		{"0.1000000000", 10},
		{"0.100000000", 9},
		{"0.10000000", 8},
		{"0.1000000", 7},
		{"0.100000", 6},
		{"0.10000", 5},
		{"0.1000", 4},
		{"0.100", 3},
		{"0.10", 2},
		{"0.1", 1},
		{"1", 1},
		{"10", 2},
		{"100", 3},
		{"1000", 4},
		{"10000", 5},
		{"100000", 6},
		{"1000000", 7},
		{"10000000", 8},
		{"100000000", 9},
		{"1000000000", 10},
		{"10000000000", 11},
		{"100000000000", 12},
		{"1000000000000", 13},
		{"10000000000000", 14},
		{"100000000000000", 15},
		{"1000000000000000", 16},
		{"10000000000000000", 17},
		{"100000000000000000", 18},
		{"1000000000000000000", 19},
	}
	for _, tt := range tests {
		d := MustParse(tt.d)
		got := d.Prec()
		if got != tt.want {
			t.Errorf("%q.Prec() = %v, want %v", tt.d, got, tt.want)
		}
	}
}

func TestDecimal_Abs(t *testing.T) {
	tests := []struct {
		d, want string
	}{
		{"1", "1"},
		{"-1", "1"},
		{"1.00", "1.00"},
		{"-1.00", "1.00"},
		{"0", "0"},
		{"0.0", "0.0"},
		{"0.00", "0.00"},
	}
	for _, tt := range tests {
		d := MustParse(tt.d)
		got := d.Abs()
		want := MustParse(tt.want)
		if got != want {
			t.Errorf("%q.Abs() = %q, want %q", d, got, want)
		}
	}
}

func TestDecimal_Neg(t *testing.T) {
	tests := []struct {
		d, want string
	}{
		{"1", "-1"},
		{"-1", "1"},
		{"1.00", "-1.00"},
		{"-1.00", "1.00"},
		{"0", "0"},
		{"0.0", "0.0"},
		{"0.00", "0.00"},
	}
	for _, tt := range tests {
		d := MustParse(tt.d)
		got := d.Neg()
		want := MustParse(tt.want)
		if got != want {
			t.Errorf("%q.Neg() = %q, want %q", d, got, want)
		}
	}
}

func TestDecimal_Predicates(t *testing.T) {
	tests := []struct {
		d                           string
		sign                        int
		isZero, isNeg, isPos, isInt bool
	}{
		{"0", 0, true, false, false, true},
		{"0.00", 0, true, false, false, true},
		{"-0.00", 0, true, false, false, true},
		{"1", 1, false, false, true, true},
		{"-1", -1, false, true, false, true},
		{"1.0", 1, false, false, true, true},
		{"1.50", 1, false, false, true, false},
		{"-1.50", -1, false, true, false, false},
		{"0.0000000000000000001", 1, false, false, true, false},
		{"-0.0000000000000000001", -1, false, true, false, false},
		{"100.000", 1, false, false, true, true},
		{"9999999999999999999", 1, false, false, true, true},
	}
	for _, tt := range tests {
		d := MustParse(tt.d)
		if got := d.Sign(); got != tt.sign {
			t.Errorf("%q.Sign() = %v, want %v", tt.d, got, tt.sign)
		}
		if got := d.IsZero(); got != tt.isZero {
			t.Errorf("%q.IsZero() = %v, want %v", tt.d, got, tt.isZero)
		}
		if got := d.IsNeg(); got != tt.isNeg {
			t.Errorf("%q.IsNeg() = %v, want %v", tt.d, got, tt.isNeg)
		}
		if got := d.IsPos(); got != tt.isPos {
			t.Errorf("%q.IsPos() = %v, want %v", tt.d, got, tt.isPos)
		}
		if got := d.IsInt(); got != tt.isInt {
			t.Errorf("%q.IsInt() = %v, want %v", tt.d, got, tt.isInt)
		}
	}
}

func TestDecimal_Scale(t *testing.T) {
	tests := []struct {
		d    string
		want int
	}{
		{"0", 0}, {"0.0", 1}, {"1.50", 2}, {"1.5", 1}, {"1e3", 0}, {"1e-5", 5},
		{"0.0000000000000000001", 19},
	}
	for _, tt := range tests {
		if got := MustParse(tt.d).Scale(); got != tt.want {
			t.Errorf("%q.Scale() = %v, want %v", tt.d, got, tt.want)
		}
	}
}

func TestDecimal_trim(t *testing.T) {
	tests := []struct {
		d     string
		scale int
		want  string
	}{
		{"1.500", 0, "1.5"},
		{"1.500", 2, "1.50"},
		{"1.500", 5, "1.500"},
		{"1.000", 0, "1"},
		{"0.000", 0, "0"},
		{"-1.500", 0, "-1.5"},
		{"100", 0, "100"},
	}
	for _, tt := range tests {
		got := MustParse(tt.d).trim(tt.scale)
		if got != MustParse(tt.want) {
			t.Errorf("%q.trim(%v) = %q, want %q", tt.d, tt.scale, got, tt.want)
		}
	}
}
