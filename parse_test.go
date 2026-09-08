package money

import (
	"testing"
)

func TestParseAmount(t *testing.T) {
	tests := []struct {
		name      string
		tok       string
		exp       int
		wantUnits int64
		wantOff   int
		wantMsg   string
	}{
		{name: "plain", tok: "19.99", exp: 2, wantUnits: 1999},
		{name: "negative", tok: "-19.99", exp: 2, wantUnits: -1999},
		{name: "no fraction needed", tok: "1500", exp: 0, wantUnits: 1500},
		{name: "pads short fraction", tok: "19.9", exp: 2, wantUnits: 1990},
		{name: "three decimal currency", tok: "1.234", exp: 3, wantUnits: 1234},
		{name: "thousands separator", tok: "5,000.00", exp: 2, wantUnits: 500000},
		{name: "multiple separators", tok: "1,234,567.89", exp: 2, wantUnits: 123456789},
		{name: "zero", tok: "0.00", exp: 2, wantUnits: 0},
		{name: "negative zero", tok: "-0.00", exp: 2, wantUnits: 0},

		{name: "empty token", tok: "", exp: 2, wantOff: 0, wantMsg: "empty amount"},
		{name: "sign only", tok: "-", exp: 2, wantOff: 1, wantMsg: "expected digits after the sign"},
		{name: "no whole digits", tok: ".50", exp: 2, wantOff: 0, wantMsg: "expected at least one digit"},
		{name: "comma before any digit", tok: ",100.00", exp: 2, wantOff: 0, wantMsg: "unexpected ',' in amount"},
		{name: "letter in whole part", tok: "1a.00", exp: 2, wantOff: 1, wantMsg: `unexpected character 'a' in amount`},
		{name: "dot with no digits after", tok: "19.", exp: 2, wantOff: 2, wantMsg: "expected digits after '.'"},
		{name: "letter in fraction", tok: "19.9a", exp: 2, wantOff: 4, wantMsg: `unexpected character 'a' in amount`},
		{name: "too many decimals", tok: "12.999", exp: 2, wantOff: 5, wantMsg: "too many decimal digits: this currency allows at most 2"},
		{name: "any fraction too many for zero-decimal currency", tok: "12.34", exp: 0, wantOff: 3, wantMsg: "too many decimal digits: this currency allows at most 0"},
		{name: "too large to represent", tok: "99999999999999999999.99", exp: 2, wantOff: 0, wantMsg: "amount is too large to represent"},

		{name: "too many digits before first separator", tok: "1234,567.00", exp: 2, wantOff: 0, wantMsg: "too many digits before the first thousands separator"},
		{name: "short group after separator", tok: "1,23,000.00", exp: 2, wantOff: 2, wantMsg: "expected 3 digits between thousands separators, got 2"},
		{name: "short trailing group", tok: "1,234,56", exp: 0, wantOff: 6, wantMsg: "expected 3 digits after the last thousands separator, got 2"},
		{name: "long trailing group", tok: "1,2345", exp: 0, wantOff: 2, wantMsg: "expected 3 digits after the last thousands separator, got 4"},
		{name: "single digit group before separator is fine", tok: "1,000.00", exp: 2, wantUnits: 100000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			units, err := parseAmount(tt.tok, tt.exp)
			if tt.wantMsg == "" {
				if err != nil {
					t.Fatalf("parseAmount(%q, %d) = error %+v, want success", tt.tok, tt.exp, err)
				}
				if units != tt.wantUnits {
					t.Fatalf("parseAmount(%q, %d) = %d, want %d", tt.tok, tt.exp, units, tt.wantUnits)
				}
				return
			}
			if err == nil {
				t.Fatalf("parseAmount(%q, %d) = %d, want error %q", tt.tok, tt.exp, units, tt.wantMsg)
			}
			if err.offset != tt.wantOff {
				t.Errorf("parseAmount(%q, %d) offset = %d, want %d", tt.tok, tt.exp, err.offset, tt.wantOff)
			}
			if err.msg != tt.wantMsg {
				t.Errorf("parseAmount(%q, %d) msg = %q, want %q", tt.tok, tt.exp, err.msg, tt.wantMsg)
			}
		})
	}
}

// TestParsePositions checks that Parse reports the line and column of the
// exact character at fault, since that's the whole point of the annotated
// error and a regression here would go unnoticed by anything that only
// checks the error message.
func TestParsePositions(t *testing.T) {
	tests := []struct {
		name    string
		line    string
		wantCol int
		wantMsg string
	}{
		{name: "code too short", line: "US 19.99", wantCol: 1, wantMsg: `"US" is not a 3-letter currency code like USD or EUR`},
		{name: "code lowercase", line: "usd 19.99", wantCol: 1, wantMsg: `"usd" is not a 3-letter currency code like USD or EUR`},
		{name: "unknown code", line: "XYZ 10.00", wantCol: 1, wantMsg: `unknown currency code "XYZ"`},
		{name: "missing amount", line: "USD", wantCol: 4, wantMsg: "expected an amount after the currency code"},
		{name: "missing amount trailing space", line: "USD   ", wantCol: 7, wantMsg: "expected an amount after the currency code"},
		{name: "trailing text", line: "USD 19.99 extra", wantCol: 11, wantMsg: `unexpected trailing text "extra"`},
		{name: "amount error offset carried through", line: "USD 12.999", wantCol: 10, wantMsg: "too many decimal digits: this currency allows at most 2"},
		{name: "comma before digit in context", line: "USD ,100.00", wantCol: 5, wantMsg: "unexpected ',' in amount"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, errs := Parse(tt.line)
			if len(errs) != 1 {
				t.Fatalf("Parse(%q) returned %d errors, want 1: %v", tt.line, len(errs), errs)
			}
			e := errs[0]
			if e.Pos.Line != 1 {
				t.Errorf("Parse(%q) line = %d, want 1", tt.line, e.Pos.Line)
			}
			if e.Pos.Column != tt.wantCol {
				t.Errorf("Parse(%q) column = %d, want %d", tt.line, e.Pos.Column, tt.wantCol)
			}
			if e.Message != tt.wantMsg {
				t.Errorf("Parse(%q) message = %q, want %q", tt.line, e.Message, tt.wantMsg)
			}
		})
	}
}

func TestParseMultiLineCollectsAllErrors(t *testing.T) {
	doc := "USD 19.99\n" +
		"XYZ 10.00\n" +
		"EUR -5,000.00\n" +
		"JPY 12.34\n"

	amounts, errs := Parse(doc)

	wantAmounts := []Amount{
		{Currency: "USD", Units: 1999},
		{Currency: "EUR", Units: -500000},
	}
	if len(amounts) != len(wantAmounts) {
		t.Fatalf("got %d amounts, want %d: %v", len(amounts), len(wantAmounts), amounts)
	}
	for i, a := range amounts {
		if a != wantAmounts[i] {
			t.Errorf("amounts[%d] = %+v, want %+v", i, a, wantAmounts[i])
		}
	}

	if len(errs) != 2 {
		t.Fatalf("got %d errors, want 2: %v", len(errs), errs)
	}
	if errs[0].Pos != (Position{Line: 2, Column: 1}) {
		t.Errorf("errs[0].Pos = %v, want 2:1", errs[0].Pos)
	}
	if errs[1].Pos != (Position{Line: 4, Column: 8}) {
		t.Errorf("errs[1].Pos = %v, want 4:8", errs[1].Pos)
	}
}

func TestParseSkipsBlankLinesAndComments(t *testing.T) {
	doc := "USD 19.99\n" +
		"\n" +
		"# a comment\n" +
		"   \n" +
		"EUR 5.00\n"

	amounts, errs := Parse(doc)
	if len(errs) != 0 {
		t.Fatalf("got errors, want none: %v", errs)
	}
	want := []Amount{
		{Currency: "USD", Units: 1999},
		{Currency: "EUR", Units: 500},
	}
	if len(amounts) != len(want) {
		t.Fatalf("got %d amounts, want %d: %v", len(amounts), len(want), amounts)
	}
	for i, a := range amounts {
		if a != want[i] {
			t.Errorf("amounts[%d] = %+v, want %+v", i, a, want[i])
		}
	}
}

func TestParseHandlesCarriageReturns(t *testing.T) {
	amounts, errs := Parse("USD 19.99\r\nEUR 5.00\r\n")
	if len(errs) != 0 {
		t.Fatalf("got errors, want none: %v", errs)
	}
	if len(amounts) != 2 {
		t.Fatalf("got %d amounts, want 2: %v", len(amounts), amounts)
	}
}

func TestParseErrorAnnotated(t *testing.T) {
	_, errs := Parse("XYZ 10.00")
	if len(errs) != 1 {
		t.Fatalf("got %d errors, want 1", len(errs))
	}
	got := errs[0].Annotated()
	want := "1:1: unknown currency code \"XYZ\"\n    XYZ 10.00\n    ^"
	if got != want {
		t.Errorf("Annotated() =\n%s\nwant\n%s", got, want)
	}
}
