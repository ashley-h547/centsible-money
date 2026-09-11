package money

import (
	"strings"
	"testing"
)

func TestLoadPrecisionTable(t *testing.T) {
	src := "USD 2\n" +
		"\n" +
		"# a made up currency\n" +
		"XTS 4\n" +
		"JPY 0\n"

	table, err := LoadPrecisionTable(strings.NewReader(src))
	if err != nil {
		t.Fatalf("LoadPrecisionTable() error = %v", err)
	}

	want := map[string]int{"USD": 2, "XTS": 4, "JPY": 0}
	if len(table) != len(want) {
		t.Fatalf("got %d entries, want %d: %v", len(table), len(want), table)
	}
	for code, exp := range want {
		if table[code] != exp {
			t.Errorf("table[%q] = %d, want %d", code, table[code], exp)
		}
	}
}

func TestLoadPrecisionTableErrors(t *testing.T) {
	tests := []struct {
		name string
		src  string
	}{
		{name: "wrong field count", src: "USD 2 extra\n"},
		{name: "code not three letters", src: "US 2\n"},
		{name: "code lowercase", src: "usd 2\n"},
		{name: "digit count not a number", src: "USD two\n"},
		{name: "negative digit count", src: "USD -1\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := LoadPrecisionTable(strings.NewReader(tt.src)); err == nil {
				t.Fatalf("LoadPrecisionTable(%q) = nil error, want one", tt.src)
			}
		})
	}
}

func TestParseWithPrecisionOverridesBuiltinTable(t *testing.T) {
	table := map[string]int{"XTS": 4}

	amounts, errs := ParseWithPrecision("XTS 1.2345\n", table)
	if len(errs) != 0 {
		t.Fatalf("got errors, want none: %v", errs)
	}
	want := []Amount{{Currency: "XTS", Units: 12345}}
	if len(amounts) != 1 || amounts[0] != want[0] {
		t.Fatalf("amounts = %+v, want %+v", amounts, want)
	}

	// USD is only known through the built-in table, so it's rejected
	// once a custom table takes over.
	_, errs = ParseWithPrecision("USD 19.99\n", table)
	if len(errs) != 1 {
		t.Fatalf("got %d errors, want 1: %v", len(errs), errs)
	}
	if errs[0].Message != `unknown currency code "USD"` {
		t.Errorf("errs[0].Message = %q, want unknown currency code", errs[0].Message)
	}
}

func TestParseWithPrecisionNilFallsBackToBuiltin(t *testing.T) {
	amounts, errs := ParseWithPrecision("USD 19.99\n", nil)
	if len(errs) != 0 {
		t.Fatalf("got errors, want none: %v", errs)
	}
	want := Amount{Currency: "USD", Units: 1999}
	if len(amounts) != 1 || amounts[0] != want {
		t.Fatalf("amounts = %+v, want [%+v]", amounts, want)
	}
}
