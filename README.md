[![Go Reference](https://pkg.go.dev/badge/github.com/jackc/numfmt.svg)](https://pkg.go.dev/github.com/jackc/numfmt)
[![Build Status](https://travis-ci.org/jackc/numfmt.svg)](https://travis-ci.org/jackc/numfmt)

# numfmt

`numfmt` is a number formatting package for Go.

## Features

* Rounding to N decimal places
* Always display minimum of N decimal places
* Configurable thousands separators
* Scaling for percentage formatting
* Format negative values differently for correct currency output like `-$12.34` or `(12.34)`

## Example

```go
f := &numfmt.Formatter{
  NegativeTemplate: "(n)",
  MinDecimalPlaces: 2,
}
f.Format("-1234") // => "(1,234.00)"
```

See the [documentation](https://pkg.go.dev/github.com/jackc/numfmt) for more examples.
