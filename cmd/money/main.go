// Command money reads a file of currency amounts, one per line, and
// prints the total for each currency. If any line fails to parse, it
// prints every error it found, each with a line, a column, and a caret
// pointing at the mistake, instead of stopping at the first one.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	money "github.com/ashley-h547/centsible-money"
)

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// jsonTotal is the shape of one entry in --json output. Currency is
// omitted under --sum-only, since the point of that flag is to drop
// the currency label from the result entirely.
type jsonTotal struct {
	Currency  string `json:"currency,omitempty"`
	Units     int64  `json:"units"`
	Formatted string `json:"formatted"`
}

func run(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("money", flag.ContinueOnError)
	fs.SetOutput(stderr)
	precisionPath := fs.String("precision", "", "path to a currency precision table file, overriding the built-in one")
	sumOnly := fs.Bool("sum-only", false, "print only the numeric total for each currency, without the currency code")
	jsonOutput := fs.Bool("json", false, "print totals as a JSON array instead of plain text")
	if err := fs.Parse(args); err != nil {
		return err
	}

	precision, err := loadPrecision(*precisionPath)
	if err != nil {
		return err
	}

	var r io.Reader = os.Stdin
	if fs.NArg() > 0 {
		f, err := os.Open(fs.Arg(0))
		if err != nil {
			return err
		}
		defer f.Close()
		r = f
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("reading input: %w", err)
	}

	amounts, errs := money.ParseWithPrecision(string(data), precision)
	if len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintln(stderr, e.Annotated())
		}
		return fmt.Errorf("%d error(s) found", len(errs))
	}

	totals := map[string]int64{}
	var order []string
	for _, a := range amounts {
		if _, seen := totals[a.Currency]; !seen {
			order = append(order, a.Currency)
		}
		totals[a.Currency] += a.Units
	}

	if *jsonOutput {
		results := make([]jsonTotal, 0, len(order))
		for _, code := range order {
			formatted := money.FormatAmount(money.Amount{Currency: code, Units: totals[code]}, precision)
			t := jsonTotal{Units: totals[code], Formatted: formatted}
			if *sumOnly {
				t.Formatted = strings.TrimPrefix(formatted, code+" ")
			} else {
				t.Currency = code
			}
			results = append(results, t)
		}
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(results)
	}

	for _, code := range order {
		formatted := money.FormatAmount(money.Amount{Currency: code, Units: totals[code]}, precision)
		if *sumOnly {
			formatted = strings.TrimPrefix(formatted, code+" ")
		}
		fmt.Fprintln(stdout, formatted)
	}
	return nil
}

// loadPrecision reads the currency precision table at path, or returns nil
// (telling Parse to fall back to its built-in table) if path is empty.
func loadPrecision(path string) (map[string]int, error) {
	if path == "" {
		return nil, nil
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening precision table: %w", err)
	}
	defer f.Close()

	table, err := money.LoadPrecisionTable(f)
	if err != nil {
		return nil, fmt.Errorf("loading precision table: %w", err)
	}
	return table, nil
}
