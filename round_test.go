package decimal

import (
	"errors"
	"testing"
)

// modes lists the modes in the column order of the tables below.
var modes = [6]RoundingMode{Floor, Ceil, Down, Up, HalfUp, HalfEven}

func TestRoundingMode_String(t *testing.T) {
	want := [6]string{"Floor", "Ceil", "Down", "Up", "HalfUp", "HalfEven"}
	for i, m := range modes {
		if got := m.String(); got != want[i] {
			t.Errorf("RoundingMode(%d).String() = %q, want %q", i, got, want[i])
		}
	}
	if got := RoundingMode(9).String(); got != "RoundingMode(9)" {
		t.Errorf("RoundingMode(9).String() = %q", got)
	}
}

func TestDecimal_Round(t *testing.T) {
	// want columns: Floor, Ceil, Down, Up, HalfUp, HalfEven
	tests := []struct {
		d     string
		scale int
		want  [6]string
	}{
		{"2.5", 0, [6]string{"2", "3", "2", "3", "3", "2"}},
		{"-2.5", 0, [6]string{"-3", "-2", "-2", "-3", "-3", "-2"}},
		{"3.5", 0, [6]string{"3", "4", "3", "4", "4", "4"}},
		{"-3.5", 0, [6]string{"-4", "-3", "-3", "-4", "-4", "-4"}},
		{"1.5", 0, [6]string{"1", "2", "1", "2", "2", "2"}},
		{"0.125", 2, [6]string{"0.12", "0.13", "0.12", "0.13", "0.13", "0.12"}},
		{"-0.125", 2, [6]string{"-0.13", "-0.12", "-0.12", "-0.13", "-0.13", "-0.12"}},
		{"0.135", 2, [6]string{"0.13", "0.14", "0.13", "0.14", "0.14", "0.14"}},
		{"0.4", 0, [6]string{"0", "1", "0", "1", "0", "0"}},
		{"-0.4", 0, [6]string{"-1", "0", "0", "-1", "0", "0"}},
		{"0.5", 0, [6]string{"0", "1", "0", "1", "1", "0"}},
		{"-0.5", 0, [6]string{"-1", "0", "0", "-1", "-1", "0"}},
		{"1.23", 0, [6]string{"1", "2", "1", "2", "1", "1"}},

		// Already rounded: the scale is never increased
		{"1.23", 2, [6]string{"1.23", "1.23", "1.23", "1.23", "1.23", "1.23"}},
		{"1.23", 5, [6]string{"1.23", "1.23", "1.23", "1.23", "1.23", "1.23"}},
		{"1.2300", 5, [6]string{"1.2300", "1.2300", "1.2300", "1.2300", "1.2300", "1.2300"}},
		{"7", 0, [6]string{"7", "7", "7", "7", "7", "7"}},

		// Trailing zeros are dropped down to the scale
		{"1.2300", 2, [6]string{"1.23", "1.23", "1.23", "1.23", "1.23", "1.23"}},
		{"100.000", 0, [6]string{"100", "100", "100", "100", "100", "100"}},
		{"0.000", 1, [6]string{"0.0", "0.0", "0.0", "0.0", "0.0", "0.0"}},

		// 19-digit coefficients
		{"999999999999999999.9", 0, [6]string{"999999999999999999", "1000000000000000000", "999999999999999999", "1000000000000000000", "1000000000000000000", "1000000000000000000"}},
		{"9999999999999999.999", 1, [6]string{"9999999999999999.9", "10000000000000000.0", "9999999999999999.9", "10000000000000000.0", "10000000000000000.0", "10000000000000000.0"}},
		{"-999999999999999999.9", 0, [6]string{"-1000000000000000000", "-999999999999999999", "-999999999999999999", "-1000000000000000000", "-1000000000000000000", "-1000000000000000000"}},
		{"0.9999999999999999999", 0, [6]string{"0", "1", "0", "1", "1", "1"}},
		{"-0.9999999999999999999", 0, [6]string{"-1", "0", "0", "-1", "-1", "-1"}},
		{"0.9999999999999999999", 18, [6]string{"0.999999999999999999", "1.000000000000000000", "0.999999999999999999", "1.000000000000000000", "1.000000000000000000", "1.000000000000000000"}},
		{"0.0000000000000000001", 18, [6]string{"0.000000000000000000", "0.000000000000000001", "0.000000000000000000", "0.000000000000000001", "0.000000000000000000", "0.000000000000000000"}},
		{"-0.0000000000000000001", 18, [6]string{"-0.000000000000000001", "0.000000000000000000", "0.000000000000000000", "-0.000000000000000001", "0.000000000000000000", "0.000000000000000000"}},
		{"9999999999999999999", 0, [6]string{"9999999999999999999", "9999999999999999999", "9999999999999999999", "9999999999999999999", "9999999999999999999", "9999999999999999999"}},

		// Scale clamping
		{"1.5", -1, [6]string{"1", "2", "1", "2", "2", "2"}},
		{"1.5", -100, [6]string{"1", "2", "1", "2", "2", "2"}},
		{"1.5", 100, [6]string{"1.5", "1.5", "1.5", "1.5", "1.5", "1.5"}},
	}
	for _, tt := range tests {
		d := MustParse(tt.d)
		for i, mode := range modes {
			got := d.Round(tt.scale, mode)
			want := MustParse(tt.want[i])
			if got != want {
				t.Errorf("%q.Round(%v, %v) = %q, want %q", d, tt.scale, mode, got, want)
			}
		}
	}
}

func TestDecimal_Round_invalidMode(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Errorf("Round with an invalid mode did not panic")
		}
	}()
	MustParse("1.5").Round(0, RoundingMode(6))
}

func TestDecimal_Rescale(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			d     string
			scale int
			want  string
		}{
			{"1.5", 2, "1.50"},
			{"1.5", 1, "1.5"},
			{"1", 3, "1.000"},
			{"-1.5", 3, "-1.500"},
			{"0", 19, "0.0000000000000000000"},
			{"1.50", 1, "1.5"},
			{"100.00", 0, "100"},
			{"-1.50", 1, "-1.5"},
			{"0.00", 0, "0"},
			{"999999999999999999", 1, "999999999999999999.0"},
			{"0.9999999999999999990", 18, "0.999999999999999999"},
		}
		for _, tt := range tests {
			d := MustParse(tt.d)
			got, err := d.Rescale(tt.scale)
			if err != nil {
				t.Errorf("%q.Rescale(%v) failed: %v", d, tt.scale, err)
				continue
			}
			if want := MustParse(tt.want); got != want {
				t.Errorf("%q.Rescale(%v) = %q, want %q", d, tt.scale, got, want)
			}
		}
	})

	t.Run("error", func(t *testing.T) {
		tests := []struct {
			d     string
			scale int
			want  error
		}{
			{"1.50", 0, ErrInexact},
			{"0.001", 2, ErrInexact},
			{"-1.55", 1, ErrInexact},
			{"1", 19, ErrOverflow},
			{"9999999999999999999", 1, ErrOverflow},
			{"1", -1, ErrOverflow},
			{"1", 20, ErrOverflow},
		}
		for _, tt := range tests {
			d := MustParse(tt.d)
			_, err := d.Rescale(tt.scale)
			if !errors.Is(err, tt.want) {
				t.Errorf("%q.Rescale(%v) error = %v, want %v", d, tt.scale, err, tt.want)
			}
		}
	})
}

func TestDecimal_Trim(t *testing.T) {
	tests := []struct{ d, want string }{
		{"1.500", "1.5"},
		{"100.00", "100"},
		{"0.00", "0"},
		{"-1.0", "-1"},
		{"1.05", "1.05"},
		{"100", "100"},
		{"0.0000000000000000010", "0.000000000000000001"},
	}
	for _, tt := range tests {
		got := MustParse(tt.d).Trim()
		if want := MustParse(tt.want); got != want {
			t.Errorf("%q.Trim() = %q, want %q", tt.d, got, want)
		}
	}
}

func TestDecimal_RoundStep(t *testing.T) {
	tests := []struct {
		d, step string
		mode    RoundingMode
		want    string
	}{
		// step 1
		{"2.5", "1", HalfEven, "2"},
		{"2.5", "1", HalfUp, "3"},
		{"2.5", "1", Floor, "2"},
		{"2.5", "1", Ceil, "3"},
		{"-2.5", "1", Floor, "-3"},
		{"-2.5", "1", Ceil, "-2"},
		{"-2.5", "1", Down, "-2"},
		{"-2.5", "1", Up, "-3"},
		{"-2.5", "1", HalfUp, "-3"},
		{"-2.5", "1", HalfEven, "-2"},

		// step 0.1
		{"65000.17", "0.1", HalfUp, "65000.2"},
		{"65000.17", "0.1", Floor, "65000.1"},
		{"65000.17", "0.1", Ceil, "65000.2"},
		{"65000.15", "0.1", HalfEven, "65000.2"},
		{"65000.25", "0.1", HalfEven, "65000.2"},
		{"0.25", "0.1", HalfUp, "0.3"},
		{"0.35", "0.1", HalfEven, "0.4"},
		{"-0.25", "0.1", HalfUp, "-0.3"},
		{"-0.25", "0.1", HalfEven, "-0.2"},
		{"-0.04", "0.1", HalfUp, "0.0"},
		{"-0.04", "0.1", Floor, "-0.1"},

		// step 0.0001
		{"1.23456", "0.0001", HalfEven, "1.2346"},
		{"1.23455", "0.0001", HalfEven, "1.2346"},
		{"1.23445", "0.0001", HalfEven, "1.2344"},
		{"1.23445", "0.0001", HalfUp, "1.2345"},
		{"0.01234", "0.001", Floor, "0.012"},
		{"0.01234", "0.001", Ceil, "0.013"},

		// step 0.25 (result scale 2)
		{"1.1", "0.25", Floor, "1.00"},
		{"1.1", "0.25", Ceil, "1.25"},
		{"1.1", "0.25", HalfUp, "1.00"},
		{"1.125", "0.25", HalfUp, "1.25"},
		{"1.125", "0.25", HalfEven, "1.00"},
		{"-1.125", "0.25", HalfUp, "-1.25"},
		{"-1.125", "0.25", HalfEven, "-1.00"},

		// step 0.5
		{"2.25", "0.5", HalfEven, "2.0"},
		{"2.25", "0.5", HalfUp, "2.5"},
		{"2.75", "0.5", HalfEven, "3.0"},
		{"0.3", "0.5", Up, "0.5"},
		{"-0.3", "0.5", Up, "-0.5"},
		{"-0.3", "0.5", Down, "0.0"},

		// step 5
		{"12", "5", Floor, "10"},
		{"12", "5", Ceil, "15"},
		{"12", "5", HalfUp, "10"},
		{"12.5", "5", HalfUp, "15"},
		{"12.5", "5", HalfEven, "10"},
		{"-12.5", "5", HalfUp, "-15"},
		{"-12.5", "5", HalfEven, "-10"},
		{"17.5", "5", HalfEven, "20"},

		// step 0.00000001
		{"0.123456785", "0.00000001", HalfEven, "0.12345678"},
		{"0.123456785", "0.00000001", HalfUp, "0.12345679"},
		{"0.123456785", "0.00000001", Floor, "0.12345678"},
		{"0.123456785", "0.00000001", Ceil, "0.12345679"},

		// Already on the step: only the scale changes
		{"1.5", "0.01", HalfUp, "1.50"},
		{"1.5", "0.01", Floor, "1.50"},
		{"-1.5", "0.01", Ceil, "-1.50"},
		{"0", "0.5", HalfEven, "0.0"},
		{"1.50", "0.5", HalfEven, "1.5"},
		{"1.500", "1", Floor, "1"},
		{"-0.1", "1", Ceil, "0"},
		{"0.1", "1", Floor, "0"},

		// Large values: the aligned coefficient needs *big.Int
		{"0.9999999999999999999", "1", HalfUp, "1"},
		{"0.9999999999999999999", "1", Floor, "0"},
		{"-0.9999999999999999999", "1", Floor, "-1"},
		{"0.4999999999999999999", "1", HalfUp, "0"},
		{"0.5000000000000000000", "1", HalfEven, "0"},
		{"0.5000000000000000001", "1", HalfEven, "1"},
		{"5000000000000000000", "3", Floor, "4999999999999999998"},
		{"9999999999999999999", "9999999999999999999", Floor, "9999999999999999999"},
		{"9999999999999999998", "1", HalfEven, "9999999999999999998"},
		{"9999999999999999999", "1", Up, "9999999999999999999"},
	}
	for _, tt := range tests {
		d := MustParse(tt.d)
		step := MustParse(tt.step)
		got, err := d.RoundStep(step, tt.mode)
		if err != nil {
			t.Errorf("%q.RoundStep(%q, %v) failed: %v", d, step, tt.mode, err)
			continue
		}
		want := MustParse(tt.want)
		if got != want {
			t.Errorf("%q.RoundStep(%q, %v) = %q, want %q", d, step, tt.mode, got, want)
		}
		if got.Scale() != step.Scale() {
			t.Errorf("%q.RoundStep(%q, %v) scale = %v, want %v", d, step, tt.mode, got.Scale(), step.Scale())
		}
	}

	t.Run("error", func(t *testing.T) {
		tests := []struct {
			d, step string
			mode    RoundingMode
			want    error
		}{
			{"1", "0", HalfUp, ErrInvalidStep},
			{"1", "0.00", HalfUp, ErrInvalidStep},
			{"1", "-1", HalfUp, ErrInvalidStep},
			{"1", "-0.5", Floor, ErrInvalidStep},
			{"1", "1", RoundingMode(6), ErrInvalidMode},
			{"9999999999999999999", "2", Ceil, ErrOverflow},
			{"9999999999999999999", "0.1", Floor, ErrOverflow},
			{"-9999999999999999999", "2", Floor, ErrOverflow},
		}
		for _, tt := range tests {
			d := MustParse(tt.d)
			step := MustParse(tt.step)
			_, err := d.RoundStep(step, tt.mode)
			if !errors.Is(err, tt.want) {
				t.Errorf("%q.RoundStep(%q, %v) error = %v, want %v", d, step, tt.mode, err, tt.want)
			}
		}
	})
}

func TestDecimal_IsMultipleOf(t *testing.T) {
	tests := []struct {
		d, step string
		want    bool
	}{
		{"1.5", "0.5", true},
		{"1.5", "0.25", true},
		{"1.5", "1", false},
		{"0", "0.01", true},
		{"-1.5", "0.5", true},
		{"-1.5", "1", false},
		{"1.50", "0.5", true},
		{"65000.17", "0.01", true},
		{"65000.17", "0.1", false},
		{"10", "5", true},
		{"10", "3", false},
		{"0.3", "0.1", true},

		// Invalid steps
		{"1", "0", false},
		{"0", "0", false},
		{"1", "-1", false},
		{"-1", "-1", false},

		// Large quotients and coefficients
		{"9999999999999999999", "1", true},
		{"0.9999999999999999999", "1", false},
		{"0.9999999999999999999", "0.0000000000000000001", true},
		{"9999999999999999999", "0.0000000000000000001", true},
		{"9999999999999999999", "0.0000000000000000007", false},
		{"9999999999999999999", "0.0000000000000000003", true},
	}
	for _, tt := range tests {
		d := MustParse(tt.d)
		step := MustParse(tt.step)
		if got := d.IsMultipleOf(step); got != tt.want {
			t.Errorf("%q.IsMultipleOf(%q) = %v, want %v", d, step, got, tt.want)
		}
	}
}
