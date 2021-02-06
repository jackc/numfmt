package numfmt

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

type Rounder struct {
	Places int32 // Number of decimal places to round to.
}

func (r *Rounder) Round(d decimal.Decimal) decimal.Decimal {
	return d.Round(r.Places)
}

type Formatter struct {
	GroupSeparator   string // Separator to place between groups of digits. Default: ","
	GroupSize        int    // Number of digits in a group. Default: 3
	DecimalSeparator string // Default: "."
	Rounder          *Rounder
}

func (f *Formatter) Format(v interface{}) string {
	switch v := v.(type) {
	case decimal.Decimal:
		return f.FormatDecimal(v)
	case string:
		d, err := decimal.NewFromString(v)
		if err != nil {
			return fmt.Sprint(v)
		}
		return f.FormatDecimal(d)
	default:
		s := fmt.Sprint(v)
		d, err := decimal.NewFromString(s)
		if err != nil {
			return s
		}
		return f.FormatDecimal(d)
	}
}

func (f *Formatter) FormatDecimal(d decimal.Decimal) string {
	if f.Rounder != nil {
		d = d.Round(f.Rounder.Places)
	}
	return f.formatNumberString(d.String())
}

func (f *Formatter) formatNumberString(s string) string {
	parts := strings.SplitN(s, ".", 2)
	intPart := parts[0]
	var fracPart string
	if len(parts) == 2 {
		fracPart = parts[1]
	}

	neg := false
	if intPart[0] == '-' {
		neg = true
		intPart = intPart[1:]
	}

	sb := &strings.Builder{}
	if neg {
		sb.WriteString("-")
	}

	groupSeparator := ","
	if f.GroupSeparator != "" {
		groupSeparator = f.GroupSeparator
	}
	groupSize := 3
	if f.GroupSize != 0 {
		groupSize = f.GroupSize
	}
	writeSeparateGroups(sb, intPart, groupSeparator, groupSize)

	decimalSeparator := "."
	if f.DecimalSeparator != "" {
		decimalSeparator = f.DecimalSeparator
	}
	if len(fracPart) != 0 {
		sb.WriteString(decimalSeparator)
		sb.WriteString(fracPart)
	}

	return sb.String()
}

func writeSeparateGroups(sb *strings.Builder, num, groupSeparator string, groupSize int) {
	if len(groupSeparator) == 0 || groupSize == 0 || len(num) <= groupSize {
		sb.WriteString(num)
		return
	}

	sepCount := len(num) / groupSize
	numIdx := len(num) % groupSize
	if numIdx == 0 {
		numIdx = groupSize
		sepCount--
	}
	sb.WriteString(num[:numIdx])

	for i := 0; i < sepCount; i++ {
		sb.WriteString(groupSeparator)
		lastNumIdx := numIdx
		numIdx += groupSize
		sb.WriteString(num[lastNumIdx:numIdx])
	}
}
