package decimal

import "testing"

var (
	benchSink    Decimal
	benchStrSink string
	benchIntSink int
)

func BenchmarkParse(b *testing.B) {
	for _, s := range []string{"65000.1", "0.012"} {
		b.Run(s, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				benchSink, _ = Parse(s)
			}
		})
	}
}

func BenchmarkString(b *testing.B) {
	d := MustParse("65000.17")
	b.ReportAllocs()
	for b.Loop() {
		benchStrSink = d.String()
	}
}

func BenchmarkAdd(b *testing.B) {
	d, e := MustParse("65000.17"), MustParse("0.012")
	b.ReportAllocs()
	for b.Loop() {
		benchSink, _ = d.Add(e)
	}
}

func BenchmarkMul(b *testing.B) {
	d, e := MustParse("65000.17"), MustParse("0.012")
	b.ReportAllocs()
	for b.Loop() {
		benchSink, _ = d.Mul(e)
	}
}

func BenchmarkQuo(b *testing.B) {
	d, e := MustParse("65000.17"), MustParse("0.012")
	b.Run("exact", func(b *testing.B) {
		x := MustParse("65000.17")
		y := MustParse("0.01")
		b.ReportAllocs()
		for b.Loop() {
			benchSink, _ = x.Quo(y)
		}
	})
	b.Run("inexact", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchSink, _ = d.Quo(e)
		}
	})
}

func BenchmarkQuoRound(b *testing.B) {
	d, e := MustParse("65000.17"), MustParse("3")
	b.Run("fast/HalfEven", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchSink, _ = d.QuoRound(e, 8, HalfEven)
		}
	})
	b.Run("fast/Floor", func(b *testing.B) {
		x, y := MustParse("1"), MustParse("3")
		b.ReportAllocs()
		for b.Loop() {
			benchSink, _ = x.QuoRound(y, 19, Floor)
		}
	})
	b.Run("bint", func(b *testing.B) {
		x, y := MustParse("0.9999999999999999999"), MustParse("7")
		b.ReportAllocs()
		for b.Loop() {
			benchSink, _ = x.QuoRound(y, 19, HalfEven)
		}
	})
}

func BenchmarkMarshalJSON(b *testing.B) {
	d := MustParse("65000.17")
	b.ReportAllocs()
	for b.Loop() {
		_, _ = d.MarshalJSON()
	}
}

func BenchmarkUnmarshalJSON(b *testing.B) {
	data := []byte(`65000.17`)
	var d Decimal
	b.ReportAllocs()
	for b.Loop() {
		_ = d.UnmarshalJSON(data)
	}
}

func BenchmarkCmp(b *testing.B) {
	d, e := MustParse("65000.17"), MustParse("0.012")
	b.ReportAllocs()
	for b.Loop() {
		benchIntSink = d.Cmp(e)
	}
}

func BenchmarkRoundStep(b *testing.B) {
	b.Run("65000.17/0.1/HalfUp", func(b *testing.B) {
		d, step := MustParse("65000.17"), MustParse("0.1")
		b.ReportAllocs()
		for b.Loop() {
			benchSink, _ = d.RoundStep(step, HalfUp)
		}
	})
	b.Run("0.01234/0.001/Floor", func(b *testing.B) {
		d, step := MustParse("0.01234"), MustParse("0.001")
		b.ReportAllocs()
		for b.Loop() {
			benchSink, _ = d.RoundStep(step, Floor)
		}
	})
}

func TestAllocs(t *testing.T) {
	d, e := MustParse("65000.17"), MustParse("0.012")
	step := MustParse("0.1")
	step2 := MustParse("0.001")
	small := MustParse("0.01234")
	three := MustParse("3")
	buf := make([]byte, 0, 32)
	num := []byte(`65000.17`)
	var dst Decimal
	tests := []struct {
		name string
		max  float64
		fn   func()
	}{
		{"QuoRound", 0, func() { benchSink, _ = d.QuoRound(three, 8, HalfEven) }},
		{"QuoRound19", 0, func() { benchSink, _ = MustParse("1").QuoRound(three, 19, Floor) }},
		{"AppendText", 0, func() { buf, _ = d.AppendText(buf[:0]) }},
		{"UnmarshalJSON", 0, func() { _ = dst.UnmarshalJSON(num) }},
		{"UnmarshalText", 0, func() { _ = dst.UnmarshalText(num) }},
		{"Min", 0, func() { benchSink = Min(d, e) }},
		{"Max", 0, func() { benchSink = Max(d, e) }},
		{"Add", 0, func() { benchSink, _ = d.Add(e) }},
		{"Sub", 0, func() { benchSink, _ = d.Sub(e) }},
		{"Mul", 0, func() { benchSink, _ = d.Mul(e) }},
		{"Cmp", 0, func() { benchIntSink = d.Cmp(e) }},
		{"RoundStep", 0, func() { benchSink, _ = d.RoundStep(step, HalfUp) }},
		{"RoundStep2", 0, func() { benchSink, _ = small.RoundStep(step2, Floor) }},
		{"Round", 0, func() { benchSink = d.Round(1, HalfEven) }},
		{"IsMultipleOf", 0, func() { _ = d.IsMultipleOf(step) }},
		{"Parse", 0, func() { benchSink, _ = Parse("65000.1") }},
		{"Parse2", 0, func() { benchSink, _ = Parse("0.012") }},
		{"String", 1, func() { benchStrSink = d.String() }},
	}
	for _, tt := range tests {
		if got := testing.AllocsPerRun(100, tt.fn); got > tt.max {
			t.Errorf("%v: %v allocs, want <= %v", tt.name, got, tt.max)
		}
	}
}
