// Package money parses lists of currency amounts and turns them into
// exact integer values, because floats and money don't mix.
package money

import "fmt"

// minorUnits maps an ISO 4217 currency code to how many digits follow
// its decimal point. Most currencies use 2; yen-like currencies use 0
// and a few Gulf-state dinars use 3, which is exactly the kind of thing
// that silently corrupts amounts if a parser assumes 2 everywhere.
var minorUnits = map[string]int{
	"USD": 2, "EUR": 2, "GBP": 2, "CHF": 2, "CAD": 2, "AUD": 2,
	"CNY": 2, "INR": 2, "SEK": 2, "NOK": 2, "DKK": 2, "NZD": 2,
	"JPY": 0, "KRW": 0, "VND": 0, "ISK": 0,
	"KWD": 3, "BHD": 3, "OMR": 3, "JOD": 3, "TND": 3,
}

// Amount is a currency value stored as an integer number of minor units
// (cents, pence, fils, whatever the currency calls them) so that adding
// two amounts together never loses a fraction of a cent to float rounding.
type Amount struct {
	Currency string
	Units    int64
}

// String renders the amount with the correct number of decimal places
// for its currency, e.g. "USD 19.99" or "JPY 1500".
func (a Amount) String() string {
	exp, ok := minorUnits[a.Currency]
	if !ok || exp == 0 {
		return fmt.Sprintf("%s %d", a.Currency, a.Units)
	}

	u := a.Units
	sign := ""
	if u < 0 {
		sign = "-"
		u = -u
	}
	div := int64(1)
	for i := 0; i < exp; i++ {
		div *= 10
	}
	return fmt.Sprintf("%s %s%d.%0*d", a.Currency, sign, u/div, exp, u%div)
}
