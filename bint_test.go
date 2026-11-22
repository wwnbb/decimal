// Derived from github.com/govalues/decimal v0.1.36 (MIT), see LICENSE.
package decimal

import (
	"strings"
	"testing"
)

func TestBint_pow10Table(t *testing.T) {
	if len(bpow10) != 100 {
		t.Fatalf("len(bpow10) = %v, want 100", len(bpow10))
	}
	for i, got := range bpow10 {
		want := mustParseBint("1" + strings.Repeat("0", i))
		if got.cmp(want) != 0 {
			t.Errorf("bpow10[%v] = %v, want %v", i, got.string(), want.string())
		}
	}
}

func TestBint_rshHalfEven(t *testing.T) {
	cases := []struct {
		z     string
		shift int
		want  string
	}{
		// Rounding
		{"1", 0, "1"},
		{"20", 1, "2"},
		{"18", 1, "2"},
		{"15", 1, "2"},
		{"12", 1, "1"},
		{"10", 1, "1"},
		{"8", 1, "1"},
		{"5", 1, "0"},
		{"2", 1, "0"},
		{"9999999999999999999", 19, "1"},
		{"9999999999999999999", 100, "0"},

		// Large shifts
		{"0", 17, "0"},
		{"0", 18, "0"},
		{"0", 19, "0"},
		{"0", 20, "0"},
		{"0", 21, "0"},

		{"1", 17, "0"},
		{"1", 18, "0"},
		{"1", 19, "0"},
		{"1", 20, "0"},
		{"1", 21, "0"},

		{"5000000000000000000", 17, "50"},
		{"5000000000000000000", 18, "5"},
		{"5000000000000000000", 19, "0"},
		{"5000000000000000000", 20, "0"},
		{"5000000000000000000", 21, "0"},

		{"5000000000000000001", 17, "50"},
		{"5000000000000000001", 18, "5"},
		{"5000000000000000001", 19, "1"},
		{"5000000000000000001", 20, "0"},
		{"5000000000000000001", 21, "0"},

		{"9999999999999999999", 17, "100"},
		{"9999999999999999999", 18, "10"},
		{"9999999999999999999", 19, "1"},
		{"9999999999999999999", 20, "0"},
		{"9999999999999999999", 21, "0"},

		{"10000000000000000000", 17, "100"},
		{"10000000000000000000", 18, "10"},
		{"10000000000000000000", 19, "1"},
		{"10000000000000000000", 20, "0"},
		{"10000000000000000000", 21, "0"},

		{"14999999999999999999", 17, "150"},
		{"14999999999999999999", 18, "15"},
		{"14999999999999999999", 19, "1"},
		{"14999999999999999999", 20, "0"},
		{"14999999999999999999", 21, "0"},

		{"15000000000000000000", 17, "150"},
		{"15000000000000000000", 18, "15"},
		{"15000000000000000000", 19, "2"},
		{"15000000000000000000", 20, "0"},
		{"15000000000000000000", 21, "0"},

		{"18446744073709551615", 17, "184"},
		{"18446744073709551615", 18, "18"},
		{"18446744073709551615", 19, "2"},
		{"18446744073709551615", 20, "0"},
		{"18446744073709551615", 21, "0"},
	}
	for _, tt := range cases {
		got := mustParseBint(tt.z)
		got.rshHalfEven(got, tt.shift)
		want := mustParseBint(tt.want)
		if got.cmp(want) != 0 {
			t.Errorf("%v.rshHalfEven(%v) = %v, want %v", tt.z, tt.shift, got.string(), want.string())
		}
	}
}

func TestBint_lsh(t *testing.T) {
	cases := []struct {
		z     string
		shift int
		want  string
	}{
		{"0", 1, "0"},
		{"1", 1, "10"},
		{"1", 20, "100000000000000000000"},
		{"1", 100, "1" + strings.Repeat("0", 100)},
	}
	for _, tt := range cases {
		got := mustParseBint(tt.z)
		got.lsh(got, tt.shift)
		want := mustParseBint(tt.want)
		if got.cmp(want) != 0 {
			t.Errorf("%v.lsh(%v) = %v, want %v", tt.z, tt.shift, got.string(), want.string())
		}
	}
}

// TestBint_prec is upstream's table, generated instead of written out:
// 10^(n-1) and 10^n - 1 have n digits, up to 101 digits.
func TestBint_prec(t *testing.T) {
	if got := mustParseBint("0").prec(); got != 0 {
		t.Errorf("0.prec() = %v, want 0", got)
	}
	for n := 1; n <= 101; n++ {
		for _, z := range []string{"1" + strings.Repeat("0", n-1), strings.Repeat("9", n)} {
			got := mustParseBint(z).prec()
			if got != n {
				t.Errorf("%q.prec() = %v, want %v", z, got, n)
			}
		}
	}
	if got := mustParseBint("1" + strings.Repeat("0", 101)).prec(); got != 102 {
		t.Errorf("10^101.prec() = %v, want 102", got)
	}
}

func TestBint_hasPrec(t *testing.T) {
	cases := []struct {
		z    string
		prec int
		want bool
	}{
		{"0", -1, true},
		{"0", 0, true},
		{"0", 1, false},
		{"1", 0, true},
		{"1", 1, true},
		{"1", 2, false},
		{"10", 1, true},
		{"10", 2, true},
		{"10", 3, false},

		{"100000000000000000", 19, false},  // 18 digits
		{"1000000000000000000", 19, true},  // 19 digits
		{"10000000000000000000", 19, true}, // 20 digits

		{"1000000000000000000", 18, true},  // 19 digits
		{"1000000000000000000", 19, true},  // 19 digits
		{"1000000000000000000", 20, false}, // 19 digits

		{"1" + strings.Repeat("0", 98), 100, false}, // 99 digits
		{"1" + strings.Repeat("0", 99), 100, true},  // 100 digits
		{"1" + strings.Repeat("0", 100), 100, true}, // 101 digits

		{"1" + strings.Repeat("0", 99), 99, true},   // 100 digits
		{"1" + strings.Repeat("0", 99), 100, true},  // 100 digits
		{"1" + strings.Repeat("0", 99), 101, false}, // 100 digits
	}
	for _, tt := range cases {
		z := mustParseBint(tt.z)
		got := z.hasPrec(tt.prec)
		if got != tt.want {
			t.Errorf("%v.hasPrec(%v) = %v, want %v", tt.z, tt.prec, got, tt.want)
		}
	}
}
