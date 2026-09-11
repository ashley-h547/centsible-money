package money

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// LoadPrecisionTable reads a currency precision table from r, one entry
// per line: a 3-letter code, whitespace, then the number of digits that
// currency's amounts carry after the decimal point. Blank lines and
// lines starting with '#' are ignored.
//
//	USD 2
//	JPY 0
//	# a made up currency for testing
//	XTS 4
//
// The result is meant to be passed to ParseWithPrecision, which lets a
// caller support currencies the built-in minorUnits table doesn't know
// about, or override an entry in it, without recompiling this package.
func LoadPrecisionTable(r io.Reader) (map[string]int, error) {
	table := make(map[string]int)

	scanner := bufio.NewScanner(r)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) != 2 {
			return nil, fmt.Errorf("line %d: expected \"CODE DIGITS\", got %q", lineNo, line)
		}

		code := fields[0]
		if len(code) != 3 || !isUpperAlpha(code) {
			return nil, fmt.Errorf("line %d: %q is not a 3-letter currency code like USD or EUR", lineNo, code)
		}

		digits, err := strconv.Atoi(fields[1])
		if err != nil || digits < 0 {
			return nil, fmt.Errorf("line %d: %q is not a non-negative digit count", lineNo, fields[1])
		}

		table[code] = digits
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading precision table: %w", err)
	}
	return table, nil
}
