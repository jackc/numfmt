[![Go Reference](https://pkg.go.dev/badge/github.com/jackc/numfmt.svg)](https://pkg.go.dev/github.com/jackc/numfmt)
![CI](https://github.com/jackc/numfmt/workflows/CI/badge.svg)

# numfmt

`numfmt` is a number formatting package for Go.

## Features

* Rounding to N decimal places
* Always display minimum of N decimal places
* Configurable thousands separators
* Scaling for percentage formatting
* Format negative values differently for correct currency output like `-$12.34` or `(12.34)`
* Easy to use with `text/template` and `html/template`

## Examples

Use directly from Go:

```go
f := &numfmt.Formatter{
  NegativeTemplate: "(n)",
  MinDecimalPlaces: 2,
}
f.Format("-1234") // => "(1,234.00)"
```

Or in use in `text/template`:

```
{{numfmt "1234.5"}} => "1,234.5"
{{numfmt "GroupSeparator" " " "DecimalSeparator" "," "1234.5"}} => "1 234,5"
```

See the [documentation](https://pkg.go.dev/github.com/jackc/numfmt) for more examples.
