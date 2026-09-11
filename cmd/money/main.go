// Command money reads a file of currency amounts, one per line, and
// prints the total for each currency. If any line fails to parse, it
// prints every error it found, each with a line, a column, and a caret
// pointing at the mistake, instead of stopping at the first one.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	money "github.com/ashley-h547/centsible-money"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	fs := flag.NewFlagSet("money", flag.ContinueOnError)
	precisionPath := fs.String("precision", "", "path to a currency precision table file, overriding the built-in one")
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
			fmt.Fprintln(os.Stderr, e.Annotated())
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
	for _, code := range order {
		fmt.Println(money.Amount{Currency: code, Units: totals[code]})
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
