package numfmt_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/jackc/numfmt"
	"github.com/shopspring/decimal"
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
