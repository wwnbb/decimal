package decimal

import (
	"encoding"
	"encoding/json"
	"errors"
	"testing"
)

var (
	_ json.Marshaler           = Decimal{}
	_ json.Unmarshaler         = (*Decimal)(nil)
	_ encoding.TextMarshaler   = Decimal{}
	_ encoding.TextUnmarshaler = (*Decimal)(nil)
	_ encoding.TextAppender    = Decimal{}
)

type jsonOrder struct {
	Price Decimal  `json:"price"`
	Qty   Decimal  `json:"qty"`
	Fee   *Decimal `json:"fee"`
}

func TestDecimal_MarshalJSON(t *testing.T) {
	fee := MustParse("0.0010")
	o := jsonOrder{Price: MustParse("65000.10"), Qty: MustParse("-0.012"), Fee: &fee}
	got, err := json.Marshal(o)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"price":65000.10,"qty":-0.012,"fee":0.0010}`
	if string(got) != want {
		t.Errorf("json.Marshal = %s, want %s", got, want)
	}

	got, err = json.Marshal(jsonOrder{})
	if err != nil {
		t.Fatal(err)
	}
	if want := `{"price":0,"qty":0,"fee":null}`; string(got) != want {
		t.Errorf("json.Marshal(zero) = %s, want %s", got, want)
	}

	got, err = json.Marshal([]Decimal{MustParse("1"), MustParse("-0.5")})
	if err != nil || string(got) != `[1,-0.5]` {
		t.Errorf("json.Marshal(slice) = %s, %v", got, err)
	}
}

func TestDecimal_UnmarshalJSON(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct{ in, want string }{
			{`1.50`, "1.50"},
			{`"1.50"`, "1.50"},
			{`" 1.5 "`, "1.5"},
			{`-0.012`, "-0.012"},
			{`"-0.012"`, "-0.012"},
			{`1e-5`, "0.00001"},
			{`1E3`, "1000"},
			{`"1e3"`, "1000"},
			{`-0`, "0"},
			{`0`, "0"},
			{`9999999999999999999`, "9999999999999999999"},
			{` 2.5 `, "2.5"},
		}
		for _, tt := range tests {
			var d Decimal
			if err := d.UnmarshalJSON([]byte(tt.in)); err != nil {
				t.Errorf("UnmarshalJSON(%s) failed: %v", tt.in, err)
				continue
			}
			if want := MustParse(tt.want); d != want {
				t.Errorf("UnmarshalJSON(%s) = %q, want %q", tt.in, d, want)
			}
		}
	})

	t.Run("struct", func(t *testing.T) {
		for _, in := range []string{
			`{"price":65000.10,"qty":"-0.012","fee":null}`,
			`{"price":"65000.10","qty":-0.012}`,
		} {
			var o jsonOrder
			if err := json.Unmarshal([]byte(in), &o); err != nil {
				t.Errorf("json.Unmarshal(%s) failed: %v", in, err)
				continue
			}
			if o.Price != MustParse("65000.10") || o.Qty != MustParse("-0.012") || o.Fee != nil {
				t.Errorf("json.Unmarshal(%s) = %+v", in, o)
			}
		}
		var o jsonOrder
		if err := json.Unmarshal([]byte(`{"fee":"0.5"}`), &o); err != nil || o.Fee == nil || *o.Fee != MustParse("0.5") {
			t.Errorf("pointer field: %+v, %v", o, err)
		}
	})

	t.Run("round trip", func(t *testing.T) {
		for _, s := range []string{"0", "1.50", "-0.0000000000000000001", "9999999999999999999", "65000.17"} {
			in := jsonOrder{Price: MustParse(s), Qty: MustParse(s)}
			b, err := json.Marshal(in)
			if err != nil {
				t.Fatal(err)
			}
			var out jsonOrder
			if err := json.Unmarshal(b, &out); err != nil {
				t.Fatal(err)
			}
			if out != in {
				t.Errorf("round trip of %s: %+v -> %s -> %+v", s, in, b, out)
			}
		}
	})

	t.Run("error", func(t *testing.T) {
		tests := []struct {
			in   string
			want error
		}{
			{`null`, ErrSyntax},
			{` null `, ErrSyntax},
			{``, ErrSyntax},
			{`"`, ErrSyntax},
			{`"1.5`, ErrSyntax},
			{`""`, ErrSyntax},
			{`"abc"`, ErrSyntax},
			{`"NaN"`, ErrSyntax},
			{`NaN`, ErrSyntax},
			{`true`, ErrSyntax},
			{`[1]`, ErrSyntax},
			{`{}`, ErrSyntax},
			{"\"1\\u002e5\"", ErrSyntax}, // escapes in strings are not supported
			{`1e999`, ErrSyntax},
			{`1e19`, ErrOverflow},
			{`"10000000000000000000"`, ErrOverflow},
		}
		for _, tt := range tests {
			d := MustParse("7")
			err := d.UnmarshalJSON([]byte(tt.in))
			if !errors.Is(err, tt.want) {
				t.Errorf("UnmarshalJSON(%s) error = %v, want %v", tt.in, err, tt.want)
			}
			if d != MustParse("7") {
				t.Errorf("UnmarshalJSON(%s) changed d to %q on error", tt.in, d)
			}
		}

		var o jsonOrder
		err := json.Unmarshal([]byte(`{"price":null}`), &o)
		if !errors.Is(err, ErrSyntax) {
			t.Errorf("json.Unmarshal null into Decimal: %v, want ErrSyntax", err)
		}
		err = json.Unmarshal([]byte(`{"price":"abc"}`), &o)
		if !errors.Is(err, ErrSyntax) {
			t.Errorf("json.Unmarshal abc into Decimal: %v, want ErrSyntax", err)
		}
	})
}

func TestDecimal_Text(t *testing.T) {
	for _, s := range []string{"0", "1.50", "-0.0000000000000000001", "9999999999999999999", "65000.17"} {
		d := MustParse(s)
		b, err := d.MarshalText()
		if err != nil || string(b) != s {
			t.Errorf("MarshalText(%s) = %s, %v", s, b, err)
		}
		var got Decimal
		if err := got.UnmarshalText(b); err != nil || got != d {
			t.Errorf("UnmarshalText(%s) = %q, %v", b, got, err)
		}
	}

	var d Decimal
	if err := d.UnmarshalText([]byte(" 1e-5\n")); err != nil || d != MustParse("0.00001") {
		t.Errorf("UnmarshalText(' 1e-5\\n') = %q, %v", d, err)
	}
	for _, in := range []string{"", " ", "NaN", "1/3", "1e999", "null", `"1"`} {
		d := MustParse("7")
		if err := d.UnmarshalText([]byte(in)); !errors.Is(err, ErrSyntax) || d != MustParse("7") {
			t.Errorf("UnmarshalText(%q) = %v (d=%q), want ErrSyntax and d unchanged", in, err, d)
		}
	}
	if err := d.UnmarshalText([]byte("1e19")); !errors.Is(err, ErrOverflow) {
		t.Errorf("UnmarshalText(1e19) = %v, want ErrOverflow", err)
	}
}

func TestDecimal_AppendText(t *testing.T) {
	d := MustParse("-65000.170")
	b := []byte("price=")
	b, err := d.AppendText(b)
	if err != nil || string(b) != "price=-65000.170" {
		t.Errorf("AppendText = %q, %v", b, err)
	}
	b, err = Zero.AppendText(nil)
	if err != nil || string(b) != "0" {
		t.Errorf("AppendText(nil) = %q, %v", b, err)
	}
}
