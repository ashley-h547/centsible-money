package money

import (
	"fmt"
	"strconv"
	"strings"
)

// Position locates a single character in the source document using the
// same 1-based line/column convention as most compilers, so an editor
// can jump straight to it.
type Position struct {
	Line   int
	Column int
}

func (p Position) String() string {
	return fmt.Sprintf("%d:%d", p.Line, p.Column)
}

// ParseError describes exactly what was wrong with one line of input
// and exactly where, so the caller never has to guess which amount in a
// long file is the problem.
type ParseError struct {
	Pos     Position
	Message string
	Source  string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("%s: %s", e.Pos, e.Message)
}

// Annotated renders the error the way a compiler would: the message,
// the offending source line, and a caret pointing at the exact column.
func (e *ParseError) Annotated() string {
	pad := strings.Repeat(" ", e.Pos.Column-1)
	return fmt.Sprintf("%s: %s\n    %s\n    %s^", e.Pos, e.Message, e.Source, pad)
}

// ParseErrors collects every error found while parsing a document. A
// single malformed line shouldn't hide the mistakes on every other line.
type ParseErrors []*ParseError

func (es ParseErrors) Error() string {
	var b strings.Builder
	for i, e := range es {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(e.Error())
	}
	return b.String()
}

// Parse reads a document of one currency amount per line, for example:
//
//	USD 19.99
//	EUR -5,000.00
//	# a comment, ignored
//	JPY 1500
//
// Blank lines and lines starting with '#' are skipped. Parse does not
// stop at the first bad line: it returns every amount it could read
// along with every error it found, so a caller can report all of them
// at once instead of forcing the user through one-fix-at-a-time.
//
// Parse uses the built-in minorUnits table. Use ParseWithPrecision to
// supply a different one, for example one loaded with LoadPrecisionTable.
func Parse(src string) ([]Amount, ParseErrors) {
	return ParseWithPrecision(src, nil)
}

// ParseWithPrecision behaves like Parse but looks up each currency code's
// decimal precision in precision instead of the built-in table. Passing a
// nil precision falls back to the built-in table, so ParseWithPrecision(src,
// nil) and Parse(src) do exactly the same thing.
func ParseWithPrecision(src string, precision map[string]int) ([]Amount, ParseErrors) {
	if precision == nil {
		precision = minorUnits
	}

	var amounts []Amount
	var errs ParseErrors

	for i, line := range strings.Split(src, "\n") {
		lineNo := i + 1
		trimmed := strings.TrimRight(line, "\r")

		fields := strings.Fields(trimmed)
		if len(fields) == 0 || strings.HasPrefix(fields[0], "#") {
			continue
		}

		amt, err := parseLine(trimmed, lineNo, precision)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		amounts = append(amounts, amt)
	}
	return amounts, errs
}

func parseLine(line string, lineNo int, precision map[string]int) (Amount, *ParseError) {
	i := 0
	col := 1
	skipSpace := func() {
		for i < len(line) && (line[i] == ' ' || line[i] == '\t') {
			i++
			col++
		}
	}

	skipSpace()
	codeStart, codeStartIdx := col, i
	for i < len(line) && line[i] != ' ' && line[i] != '\t' {
		i++
		col++
	}
	code := line[codeStartIdx:i]

	if len(code) != 3 || !isUpperAlpha(code) {
		return Amount{}, &ParseError{
			Pos:     Position{lineNo, codeStart},
			Message: fmt.Sprintf("%q is not a 3-letter currency code like USD or EUR", code),
			Source:  line,
		}
	}
	exp, known := precision[code]
	if !known {
		return Amount{}, &ParseError{
			Pos:     Position{lineNo, codeStart},
			Message: fmt.Sprintf("unknown currency code %q", code),
			Source:  line,
		}
	}

	skipSpace()
	if i >= len(line) {
		return Amount{}, &ParseError{
			Pos:     Position{lineNo, col},
			Message: "expected an amount after the currency code",
			Source:  line,
		}
	}
	amtStart, amtStartIdx := col, i
	for i < len(line) && line[i] != ' ' && line[i] != '\t' {
		i++
		col++
	}

	units, aerr := parseAmount(line[amtStartIdx:i], exp)
	if aerr != nil {
		return Amount{}, &ParseError{
			Pos:     Position{lineNo, amtStart + aerr.offset},
			Message: aerr.msg,
			Source:  line,
		}
	}

	skipSpace()
	if i < len(line) {
		return Amount{}, &ParseError{
			Pos:     Position{lineNo, col},
			Message: fmt.Sprintf("unexpected trailing text %q", line[i:]),
			Source:  line,
		}
	}

	return Amount{Currency: code, Units: units}, nil
}

// amountError carries a byte offset relative to the start of the amount
// token, which parseLine turns into an absolute column.
type amountError struct {
	offset int
	msg    string
}

// parseAmount converts a token like "-5,000.00" into minor units. It
// rejects anything that doesn't match the currency's decimal precision
// instead of silently truncating, so "USD 12.999" is a caught mistake
// rather than a quietly wrong price.
func parseAmount(tok string, exp int) (int64, *amountError) {
	if tok == "" {
		return 0, &amountError{0, "empty amount"}
	}

	i := 0
	neg := false
	if tok[i] == '-' {
		neg = true
		i++
	}
	if i >= len(tok) {
		return 0, &amountError{i, "expected digits after the sign"}
	}

	var whole strings.Builder
	sawDigit := false
	sawComma := false
	group := 0 // digits seen since the start of the token or the last comma
	for i < len(tok) && tok[i] != '.' {
		switch c := tok[i]; {
		case c >= '0' && c <= '9':
			whole.WriteByte(c)
			sawDigit = true
			group++
		case c == ',':
			if !sawDigit {
				return 0, &amountError{i, "unexpected ',' in amount"}
			}
			if !sawComma && group > 3 {
				return 0, &amountError{i - group, "too many digits before the first thousands separator"}
			}
			if sawComma && group != 3 {
				return 0, &amountError{i - group, fmt.Sprintf("expected 3 digits between thousands separators, got %d", group)}
			}
			sawComma = true
			group = 0
		default:
			return 0, &amountError{i, fmt.Sprintf("unexpected character %q in amount", c)}
		}
		i++
	}
	if whole.Len() == 0 {
		return 0, &amountError{i, "expected at least one digit"}
	}
	if sawComma && group != 3 {
		return 0, &amountError{i - group, fmt.Sprintf("expected 3 digits after the last thousands separator, got %d", group)}
	}

	fracDigits := ""
	if i < len(tok) && tok[i] == '.' {
		dotIdx := i
		i++
		start := i
		for i < len(tok) && tok[i] >= '0' && tok[i] <= '9' {
			i++
		}
		fracDigits = tok[start:i]
		if fracDigits == "" {
			return 0, &amountError{dotIdx, "expected digits after '.'"}
		}
		if i < len(tok) {
			return 0, &amountError{i, fmt.Sprintf("unexpected character %q in amount", tok[i])}
		}
		if len(fracDigits) > exp {
			return 0, &amountError{
				start + exp,
				fmt.Sprintf("too many decimal digits: this currency allows at most %d", exp),
			}
		}
	}
	for len(fracDigits) < exp {
		fracDigits += "0"
	}

	units, err := strconv.ParseInt(whole.String()+fracDigits, 10, 64)
	if err != nil {
		return 0, &amountError{0, "amount is too large to represent"}
	}
	if neg {
		units = -units
	}
	return units, nil
}

func isUpperAlpha(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < 'A' || s[i] > 'Z' {
			return false
		}
	}
	return true
}
