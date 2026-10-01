package money

import (
	"math"
	"testing"
)

func TestAmountString(t *testing.T) {
	tests := []struct {
		amount Amount
		want   string
	}{
		{Amount{"USD", 1999}, "USD 19.99"},
		{Amount{"USD", -1999}, "USD -19.99"},
		{Amount{"JPY", 1500}, "JPY 1500"},
		{Amount{"KWD", 1234}, "KWD 1.234"},
	}
	for _, tt := range tests {
		if got := tt.amount.String(); got != tt.want {
			t.Errorf("%+v.String() = %q, want %q", tt.amount, got, tt.want)
		}
	}
}

func TestFormatAmountWithCustomPrecision(t *testing.T) {
	table := map[string]int{"XTS": 4}

	// XTS isn't in the built-in table, so String() has no way to know
	// it takes 4 decimal digits; FormatAmount needs the caller's table.
	a := Amount{Currency: "XTS", Units: 12345}
	if got, want := FormatAmount(a, table), "XTS 1.2345"; got != want {
		t.Errorf("FormatAmount(%+v, table) = %q, want %q", a, got, want)
	}

	// A custom table can also override the digit count for a code the
	// built-in table already knows.
	override := map[string]int{"USD": 3}
	u := Amount{Currency: "USD", Units: 19990}
	if got, want := FormatAmount(u, override), "USD 19.990"; got != want {
		t.Errorf("FormatAmount(%+v, override) = %q, want %q", u, got, want)
	}
}

func TestAddSub(t *testing.T) {
	a := Amount{"USD", 1999}
	b := Amount{"USD", 1}

	sum, err := a.Add(b)
	if err != nil || sum != (Amount{"USD", 2000}) {
		t.Errorf("Add = %+v, %v; want USD 2000, nil", sum, err)
	}
	diff, err := b.Sub(a)
	if err != nil || diff != (Amount{"USD", -1998}) {
		t.Errorf("Sub = %+v, %v; want USD -1998, nil", diff, err)
	}
}

func TestAddSubCurrencyMismatch(t *testing.T) {
	usd := Amount{"USD", 100}
	eur := Amount{"EUR", 100}

	if _, err := usd.Add(eur); err == nil {
		t.Error("Add of USD and EUR succeeded, want error")
	}
	if _, err := usd.Sub(eur); err == nil {
		t.Error("Sub of EUR from USD succeeded, want error")
	}
}

func TestAddSubOverflow(t *testing.T) {
	max := Amount{"USD", math.MaxInt64}
	min := Amount{"USD", math.MinInt64}
	one := Amount{"USD", 1}

	if _, err := max.Add(one); err == nil {
		t.Error("MaxInt64 + 1 succeeded, want overflow error")
	}
	if _, err := min.Add(Amount{"USD", -1}); err == nil {
		t.Error("MinInt64 + -1 succeeded, want overflow error")
	}
	if _, err := min.Sub(one); err == nil {
		t.Error("MinInt64 - 1 succeeded, want overflow error")
	}
	if _, err := max.Sub(Amount{"USD", -1}); err == nil {
		t.Error("MaxInt64 - -1 succeeded, want overflow error")
	}

	// Boundary values that still fit must not be rejected.
	if got, err := max.Add(Amount{"USD", -1}); err != nil || got.Units != math.MaxInt64-1 {
		t.Errorf("MaxInt64 + -1 = %+v, %v", got, err)
	}
	if got, err := min.Add(max); err != nil || got.Units != -1 {
		t.Errorf("MinInt64 + MaxInt64 = %+v, %v", got, err)
	}
}

func TestFormatAmountNilFallsBackToBuiltin(t *testing.T) {
	a := Amount{Currency: "USD", Units: 1999}
	if got, want := FormatAmount(a, nil), "USD 19.99"; got != want {
		t.Errorf("FormatAmount(%+v, nil) = %q, want %q", a, got, want)
	}
}
