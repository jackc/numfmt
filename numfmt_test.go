package numfmt_test

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"text/template"

	"github.com/jackc/numfmt"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

type testFormatter numfmt.Formatter

func (f *testFormatter) String() string {
	parts := []string{}
	if f.GroupSeparator != "" {
		parts = append(parts, fmt.Sprintf(`GroupSeparator: "%s"`, f.GroupSeparator))
	}
	if f.GroupSize != 0 {
		parts = append(parts, fmt.Sprintf("GroupSize: %d", f.GroupSize))
	}
	if f.DecimalSeparator != "" {
		parts = append(parts, fmt.Sprintf(`DecimalSeparator: "%s"`, f.DecimalSeparator))
	}
	if f.Rounder != nil {
		parts = append(parts, fmt.Sprintf(`Rounder: {Places: %d}`, f.Rounder.Places))
	}
	if f.Template != "" {
		parts = append(parts, fmt.Sprintf(`Template: "%s"`, f.Template))
	}
	if f.NegativeTemplate != "" {
		parts = append(parts, fmt.Sprintf(`NegativeTemplate: "%s"`, f.NegativeTemplate))
	}

	return "&Formatter{" + strings.Join(parts, ", ") + "}"
}

func TestFormatterFormat(t *testing.T) {
	for i, tt := range []struct {
		formatter *numfmt.Formatter
		arg       interface{}
		expected  string
	}{
		// Defaults
		{&numfmt.Formatter{}, "0", "0"},
		{&numfmt.Formatter{}, "1", "1"},
		{&numfmt.Formatter{}, "12", "12"},
		{&numfmt.Formatter{}, "123", "123"},
		{&numfmt.Formatter{}, "1234", "1,234"},
		{&numfmt.Formatter{}, "12345", "12,345"},
		{&numfmt.Formatter{}, "123456", "123,456"},
		{&numfmt.Formatter{}, "12345678901234567890", "12,345,678,901,234,567,890"},
		{&numfmt.Formatter{}, "1.0", "1"},
		{&numfmt.Formatter{}, "1.2", "1.2"},
		{&numfmt.Formatter{}, "12345.6789", "12,345.6789"},
		{&numfmt.Formatter{}, "-12345.6789", "-12,345.6789"},

		{&numfmt.Formatter{DecimalSeparator: ","}, "1.2", "1,2"},
		{&numfmt.Formatter{GroupSeparator: " "}, "1234", "1 234"},
		{&numfmt.Formatter{GroupSize: 1}, "1234", "1,2,3,4"},

		{&numfmt.Formatter{Rounder: &numfmt.Rounder{Places: 0}}, "1234.1", "1,234"},
		{&numfmt.Formatter{Rounder: &numfmt.Rounder{Places: 0}}, "1234.5", "1,235"},
		{&numfmt.Formatter{Rounder: &numfmt.Rounder{Places: 0}}, "1234.9", "1,235"},
		{&numfmt.Formatter{Rounder: &numfmt.Rounder{Places: 3}}, "1234.5678", "1,234.568"},
		{&numfmt.Formatter{Rounder: &numfmt.Rounder{Places: -2}}, "1234.5678", "1,200"},

		{&numfmt.Formatter{Shift: 2}, "0.31", "31"},
		{&numfmt.Formatter{Shift: -1}, "42", "4.2"},

		// Shift happens before rounding
		{&numfmt.Formatter{Shift: 2, Rounder: &numfmt.Rounder{Places: 0}}, "0.315", "32"},

		{&numfmt.Formatter{MinDecimalPlaces: 2}, "123", "123.00"},

		// Template
		{&numfmt.Formatter{Template: "hi n"}, "-1234", "hi -1234"},
		{&numfmt.Formatter{Template: "+n"}, "123", "+123"},
		{&numfmt.Formatter{Template: "-n"}, "123", "123"},
		{&numfmt.Formatter{Template: "-n"}, "-123", "-123"},
		{&numfmt.Formatter{Template: "n -"}, "-123", "123 -"},
		{&numfmt.Formatter{Template: `\n \- \+ \\ n`}, "123", `n - + \ 123`},

		// Negative Template
		{&numfmt.Formatter{NegativeTemplate: "(n)"}, "123", "123"},
		{&numfmt.Formatter{NegativeTemplate: "(n)"}, "-123", "(123)"},

		// Different argument type tests
		{&numfmt.Formatter{}, 1234, "1,234"},
		{&numfmt.Formatter{}, 1234.0, "1,234"},
		{&numfmt.Formatter{}, int32(1234), "1,234"},
		{&numfmt.Formatter{}, int64(1234), "1,234"},
		{&numfmt.Formatter{}, float32(1234.5), "1,234.5"},
		{&numfmt.Formatter{}, float64(1234.5), "1,234.5"},
		{&numfmt.Formatter{}, decimal.RequireFromString("1234"), "1,234"},

		// Not a number
		{&numfmt.Formatter{}, "foobar", "foobar"},
	} {
		actual := tt.formatter.Format(tt.arg)
		if tt.expected != actual {
			t.Errorf("%d. expected formatting %v with %v to return %v, but got %v", i, tt.arg, (*testFormatter)(tt.formatter), tt.expected, actual)
		}
	}
}

func TestTemplateFunc(t *testing.T) {
	for i, tt := range []struct {
		format   []interface{}
		arg      interface{}
		expected string
	}{
		{[]interface{}{}, "1234.5", "1,234.5"},
		{[]interface{}{"DecimalSeparator", ","}, "1.2", "1,2"},
		{[]interface{}{"GroupSeparator", " "}, "1234", "1 234"},
		{[]interface{}{"GroupSize", 1}, "1234", "1,2,3,4"},
		{[]interface{}{"RoundPlaces", 0}, "1234.9", "1,235"},
		{[]interface{}{"Shift", 2}, "0.31", "31"},
		{[]interface{}{"Shift", 2, "RoundPlaces", 0}, "0.315", "32"},
		{[]interface{}{"MinDecimalPlaces", 2}, "123", "123.00"},
		{[]interface{}{"Template", "+n"}, "123", "+123"},
		{[]interface{}{"NegativeTemplate", "(n)"}, "-123", "(123)"},
	} {
		fn, err := numfmt.TemplateFunc(tt.format...)
		assert.NoError(t, err)
		if fn, ok := fn.(func(interface{}) string); ok {
			actual := fn(tt.arg)
			if tt.expected != actual {
				t.Errorf("%d. func: expected formatting %v with %v to return %v, but got %v", i, tt.arg, tt.format, tt.expected, actual)
			}
		} else {
			t.Errorf("%d. func: expected formatting with %v to return function but did not", i, tt.format)
		}

		args := append(tt.format, tt.arg)
		actual, err := numfmt.TemplateFunc(args...)
		assert.NoError(t, err)
		if tt.expected != actual {
			t.Errorf("%d. immediate: expected formatting %v with %v to return %v, but got %v", i, tt.arg, tt.format, tt.expected, actual)
		}
	}
}

func TestNewUSDFormatter(t *testing.T) {
	for i, tt := range []struct {
		arg      interface{}
		expected string
	}{
		{"123", "$123.00"},
		{"-123", "-$123.00"},
		{"123.456", "$123.456"},
	} {
		actual := numfmt.NewUSDFormatter().Format(tt.arg)
		if tt.expected != actual {
			t.Errorf("%d. expected formatting %v to return %v, but got %v", i, tt.arg, tt.expected, actual)
		}
	}
}

func TestNewPercentFormatter(t *testing.T) {
	for i, tt := range []struct {
		arg      interface{}
		expected string
	}{
		{"0.123", "12.3%"},
		{"1.5", "150%"},
		{"-3", "-300%"},
	} {
		actual := numfmt.NewPercentFormatter().Format(tt.arg)
		if tt.expected != actual {
			t.Errorf("%d. expected formatting %v to return %v, but got %v", i, tt.arg, tt.expected, actual)
		}
	}
}

func ExampleTemplateFunc() {
	t := template.New("root").Funcs(template.FuncMap{
		"numfmt": numfmt.TemplateFunc,
	})
	t = template.Must(t.Parse(`
numfmt can be called directly:
{{numfmt "GroupSeparator" " " "DecimalSeparator" "," "1234.56789"}}
or it can return a function for later use:
{{- $formatUSD := numfmt "Template" "$n" "RoundPlaces" 2 "MinDecimalPlaces" 2}}
{{call $formatUSD "1234.56789"}}
`))

	err := t.Execute(os.Stdout, nil)
	if err != nil {
		fmt.Println(err)
	}

	// Output:
	// numfmt can be called directly:
	// 1 234,56789
	// or it can return a function for later use:
	// $1,234.57
}

func ExampleFormatter_zero() {
	f := &numfmt.Formatter{}
	fmt.Println(f.Format("1234.56789"))

	// Output:
	// 1,234.56789
}

func ExampleFormatter_rounding() {
	f := &numfmt.Formatter{
		Rounder: &numfmt.Rounder{Places: 2},
	}
	fmt.Println(f.Format("1234.56789"))

	// Output:
	// 1,234.57
}

func ExampleFormatter_negative_currency() {
	f := &numfmt.Formatter{
		NegativeTemplate: "(n)",
		MinDecimalPlaces: 2,
	}
	fmt.Println(f.Format("-1234"))

	// Output:
	// (1,234.00)
}

func ExampleNewPercentFormatter() {
	f := numfmt.NewPercentFormatter()
	fmt.Println(f.Format("0.781"))

	// Output:
	// 78.1%
}
