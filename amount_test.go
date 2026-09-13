package money

import "testing"

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

func TestFormatAmountNilFallsBackToBuiltin(t *testing.T) {
	a := Amount{Currency: "USD", Units: 1999}
	if got, want := FormatAmount(a, nil), "USD 19.99"; got != want {
		t.Errorf("FormatAmount(%+v, nil) = %q, want %q", a, got, want)
	}
}
