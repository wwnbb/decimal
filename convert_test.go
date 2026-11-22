package decimal

import (
	"errors"
	"math"
	"testing"
)

// Variables, so that f01 + f02 is a float64 sum and not an exact constant.
var f01, f02 = 0.1, 0.2

func TestNewFromFloat64(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			f    float64
			want string
		}{
			{0.1, "0.1"},
			{1e-7, "0.0000001"},
			{-2.5, "-2.5"},
			{0, "0"},
			{math.Copysign(0, -1), "0"},
			{1, "1"},
			{100, "100"},
			{123.456, "123.456"},
			{65000.1, "65000.1"},
			{f01 + f02, "0.30000000000000004"},
			{1e15 + 0.5, "1000000000000000.5"},
			{1e18, "1000000000000000000"},
			{-1e-19, "-0.0000000000000000001"},
			{123456789012345678, "123456789012345680"},

			// More than 19 digits after the point: rounded half to even like Parse
			{1e-25, "0.0000000000000000000"},
			{5e-324, "0.0000000000000000000"},
			{1.5e-19, "0.0000000000000000002"},
			{2.5e-19, "0.0000000000000000002"},
		}
		for _, tt := range tests {
			got, err := NewFromFloat64(tt.f)
			if err != nil {
				t.Errorf("NewFromFloat64(%v) failed: %v", tt.f, err)
				continue
			}
			if got.String() != tt.want {
				t.Errorf("NewFromFloat64(%v) = %q, want %q", tt.f, got, tt.want)
			}
		}
	})

	t.Run("error", func(t *testing.T) {
		tests := []struct {
			f    float64
			want error
		}{
			{math.NaN(), ErrSyntax},
			{math.Inf(1), ErrSyntax},
			{math.Inf(-1), ErrSyntax},
			{1e19, ErrOverflow},
			{-1e19, ErrOverflow},
			{1e300, ErrOverflow},
			{math.MaxFloat64, ErrOverflow},
			{-math.MaxFloat64, ErrOverflow},
		}
		for _, tt := range tests {
			_, err := NewFromFloat64(tt.f)
			if !errors.Is(err, tt.want) {
				t.Errorf("NewFromFloat64(%v) error = %v, want %v", tt.f, err, tt.want)
			}
		}
	})
}

func TestDecimal_Float64(t *testing.T) {
	tests := []struct {
		d    string
		want float64
	}{
		{"0", 0},
		{"0.1", 0.1},
		{"-2.5", -2.5},
		{"1.50", 1.5},
		{"65000.1", 65000.1},
		{"9999999999999999999", 1e19},
		{"0.0000000000000000001", 1e-19},
	}
	for _, tt := range tests {
		got, ok := MustParse(tt.d).Float64()
		if !ok || got != tt.want {
			t.Errorf("%q.Float64() = %v, %v, want %v, true", tt.d, got, ok, tt.want)
		}
	}
}

func TestFloat64_roundTrip(t *testing.T) {
	for _, f := range []float64{
		0, 0.1, -0.1, 0.5, -2.5, 1e-7, 123.456, 65000.1, f01 + f02, 1e15 + 0.5,
		1e-19, 1e18, 12345.6789, -98765.4321, 3.141592653589793, 1.0 / 3.0,
	} {
		d, err := NewFromFloat64(f)
		if err != nil {
			t.Errorf("NewFromFloat64(%v) failed: %v", f, err)
			continue
		}
		got, ok := d.Float64()
		if !ok || got != f {
			t.Errorf("Float64(NewFromFloat64(%v)) = %v, %v (decimal %q)", f, got, ok, d)
		}
	}
}

func TestDecimal_Int64(t *testing.T) {
	tests := []struct {
		d      string
		scale  int
		want   int64
		wantOk bool
	}{
		{"1.25", 2, 125, true},
		{"1.25", 1, 0, false},
		{"1.25", 0, 0, false},
		{"1.25", 3, 1250, true},
		{"-1.25", 2, -125, true},
		{"-1.25", 1, 0, false},
		{"1.50", 1, 15, true},
		{"1.50", 0, 0, false},
		{"100", 0, 100, true},
		{"100", 2, 10000, true},
		{"0", 5, 0, true},
		{"0.00", 0, 0, true},
		{"-0.00", 3, 0, true},
		{"0.001", 3, 1, true},
		{"0.001", 2, 0, false},
		{"65000.17", 2, 6500017, true},

		// Range of int64
		{"9223372036854775807", 0, math.MaxInt64, true},
		{"9223372036854775808", 0, 0, false},
		{"-9223372036854775808", 0, math.MinInt64, true},
		{"-9223372036854775809", 0, 0, false},
		{"922337203685477580.7", 1, math.MaxInt64, true},
		{"9223372036854775807", 1, 0, false},
		{"9999999999999999999", 0, 0, false},
		{"0.9999999999999999999", 19, 0, false},
		{"0.9223372036854775807", 19, math.MaxInt64, true},
		{"-0.9223372036854775808", 19, math.MinInt64, true},

		// Scale out of range
		{"1", -1, 0, false},
		{"1", 20, 0, false},
		{"0", 20, 0, false},
	}
	for _, tt := range tests {
		got, ok := MustParse(tt.d).Int64(tt.scale)
		if got != tt.want || ok != tt.wantOk {
			t.Errorf("%q.Int64(%v) = %v, %v, want %v, %v", tt.d, tt.scale, got, ok, tt.want, tt.wantOk)
		}
	}
}

func TestInt64_New_roundTrip(t *testing.T) {
	for _, v := range []int64{0, 1, -1, 125, -125, math.MaxInt64, math.MinInt64} {
		for scale := 0; scale <= MaxScale; scale++ {
			d, err := New(v, scale)
			if err != nil {
				t.Fatal(err)
			}
			got, ok := d.Int64(scale)
			if !ok || got != v {
				t.Errorf("New(%v, %v).Int64(%v) = %v, %v", v, scale, scale, got, ok)
			}
		}
	}
}
