# centsible-money

A small Go library (and a thin CLI) for parsing lists of currency
amounts. Two problems this exists to avoid:

1. Storing money as `float64` and watching cents evaporate after enough
   additions. `Amount` stores minor units (cents, pence, fils) as an
   `int64`, so `19.99 + 0.01` is exactly `20.00`, always.
2. Parsers that say "invalid input" and leave you to find the typo by
   eye in a 400-line file. This one reports a line, a column, and a
   caret under the exact character that's wrong, for every bad line at
   once, not just the first.

## The format

One amount per line: a 3-letter ISO 4217 code, whitespace, then a
number. Blank lines and lines starting with `#` are ignored.

```
USD 19.99
EUR -5,000.00
# monthly rent, converted
JPY 1500
```

Amounts must match their currency's decimal precision: two digits for
USD/EUR/GBP, zero for JPY/KRW, three for KWD/BHD/OMR, and so on. That
list lives in `amount.go` and currently covers the currencies I've
actually needed.

## Library usage

```go
package main

import (
	"fmt"

	money "github.com/ashley-h547/centsible-money"
)

func main() {
	amounts, errs := money.Parse("USD 19.99\nEUR -5,000.00\n")
	if len(errs) > 0 {
		for _, e := range errs {
			fmt.Println(e.Annotated())
		}
		return
	}
	for _, a := range amounts {
		fmt.Println(a)
	}
}
```

An error looks like this:

```
2:9: too many decimal digits: this currency allows at most 2
    USD 12.999
        ^
```

`ParseError.Error()` gives the short `line:column: message` form for
logs; `ParseError.Annotated()` gives the compiler-style version above
for anything a human is going to read.

## CLI usage

```
$ go run ./cmd/money amounts.txt
USD 45.50
EUR -5000.00
JPY 1500
```

With a mistake in the file:

```
$ go run ./cmd/money amounts.txt
3:5: unknown currency code "XYZ"
    XYZ 10.00
    ^
1 error(s) found
```

It reads from a file argument, or from stdin if no argument is given.

## Status

Early. No exchange rates, no formatting locales, no rounding modes -
just parsing amounts and adding them up correctly. See the roadmap for
what's next.

## License

MIT, see [LICENSE](LICENSE).
