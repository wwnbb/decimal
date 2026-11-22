package decimal_test

import (
	"fmt"

	"github.com/wwnbb/decimal"
)

// The example of README.md.
func Example() {
	price := decimal.MustParse("65000.17")
	qty := decimal.MustParse("0.012")

	notional, err := price.Mul(qty)
	if err != nil {
		panic(err)
	}

	tick := decimal.MustParse("0.5")
	limit, err := price.RoundStep(tick, decimal.Floor)
	if err != nil {
		panic(err)
	}

	lot := decimal.MustParse("0.001")
	ok := qty.IsMultipleOf(lot)

	fmt.Println(notional, limit, ok, notional.Round(2, decimal.HalfEven))
	// Output: 780.00204 65000.0 true 780.00
}
