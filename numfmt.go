package numfmt

import (
	"fmt"
	"strings"
	"sync"

	"github.com/shopspring/decimal"
)

type Rounder struct {
	Places int32 // Number of decimal places to round to.
}

func (r *Rounder) Round(d decimal.Decimal) decimal.Decimal {
	return d.Round(r.Places)
}

// Formatter is a formatter of numbers. The zero value is usable. Do not change or copy a Formatter after it has been
// used. The methods on Format are concurrency safe.
type Formatter struct {
	GroupSeparator   string // Separator to place between groups of digits. Default: ","
	GroupSize        int    // Number of digits in a group. Default: 3
	DecimalSeparator string // Default: "."
	Rounder          *Rounder

	// Number of places to shift decimal places to the left. Negative numbers are shifted to the right. If set to 2 this
	// will convert a fraction to a percentage.
	Shift int32

	MinDecimalPlaces int32 // Minimum number of decimal places to display.

	// Template is a simple format string. All text other than format verbs is passed through unmodified. Backslash '\'
	// escaping can be used to include a character otherwise used as a verb. If neither '-' nor '+' are in the string
	// negative numbers will be prefixed with '-' as normal.
	//
	// Verbs:
	//   n    the number
	//   -    optional negative sign
	//   +    always include sign
	//
	// Examples:
	//   "n"    => 9.45
	//   "- n"  => - 9.45
	//   "+n"   => +9.45
	//   "n +"  => 9.45 +
	//   "$n"   => $9.45
	//   "n%"   => 9.45%
	//
	// Default: "n"
	Template         string
	compiledTemplate compiledTemplate

	// NegativeTemplate will be used if present instead of Template for negative values. The primary expected use is for
	// negative values surrounded by parentheses. It uses the same verbs as Template.
	//
	// Examples:
	//   "(n)"    => (9.45)
	// Default: ""
	NegativeTemplate         string
	compiledNegativeTemplate compiledTemplate

	compileTemplateOnce sync.Once
}

// Format formats v. v can be anything that fmt.Sprint can convert to a parsable number.
func (f *Formatter) Format(v interface{}) string {
	switch v := v.(type) {
	case decimal.Decimal:
		return f.formatDecimal(v)
	case string:
		d, err := decimal.NewFromString(v)
		if err != nil {
			return fmt.Sprint(v)
		}
		return f.formatDecimal(d)
	case int32:
		return f.formatDecimal(decimal.NewFromInt32(v))
	case int64:
		return f.formatDecimal(decimal.NewFromInt(v))
	default:
		s := fmt.Sprint(v)
		d, err := decimal.NewFromString(s)
		if err != nil {
			return s
		}
		return f.formatDecimal(d)
	}
}

func (f *Formatter) formatDecimal(d decimal.Decimal) string {
	f.compileTemplateOnce.Do(f.compileTemplates)

	if f.Shift != 0 {
		d = d.Shift(f.Shift)
	}
	if f.Rounder != nil {
		d = d.Round(f.Rounder.Places)
	}

	parts := strings.SplitN(d.String(), ".", 2)
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

	if len(fracPart) < int(f.MinDecimalPlaces) {
		buf := make([]byte, int(f.MinDecimalPlaces))
		copy(buf, fracPart)
		for i := len(fracPart); i < len(buf); i++ {
			buf[i] = '0'
		}
		fracPart = string(buf)
	}

	sb := &strings.Builder{}
	if neg && f.compiledNegativeTemplate != nil {
		f.compiledNegativeTemplate.write(sb, f, neg, intPart, fracPart)
	} else {
		f.compiledTemplate.write(sb, f, neg, intPart, fracPart)
	}

	return sb.String()
}

func (f *Formatter) compileTemplates() {
	if f.compiledTemplate != nil {
		return
	}

	t := "n"
	if f.Template != "" {
		t = f.Template
	}
	f.compiledTemplate = compileTemplate(t)

	if f.NegativeTemplate == "" {
		return
	}

	f.compiledNegativeTemplate = compileTemplate(f.NegativeTemplate)
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

type compiledTemplatePart interface {
	write(sb *strings.Builder, f *Formatter, neg bool, intPart, fracPart string)
}

type compiledTemplate []compiledTemplatePart

func (ct compiledTemplate) write(sb *strings.Builder, f *Formatter, neg bool, intPart, fracPart string) {
	for _, part := range ct {
		part.write(sb, f, neg, intPart, fracPart)
	}
}

type compiledTemplatePartLiteral string

func (p compiledTemplatePartLiteral) write(sb *strings.Builder, f *Formatter, neg bool, intPart, fracPart string) {
	sb.WriteString(string(p))
}

type compiledTemplatePartNumber struct{}

func (compiledTemplatePartNumber) write(sb *strings.Builder, f *Formatter, neg bool, intPart, fracPart string) {
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
}

type compiledTemplatePartOptionalSign struct{}

func (compiledTemplatePartOptionalSign) write(sb *strings.Builder, f *Formatter, neg bool, intPart, fracPart string) {
	if neg {
		sb.WriteByte('-')
	}
}

type compiledTemplatePartForceSign struct{}

func (compiledTemplatePartForceSign) write(sb *strings.Builder, f *Formatter, neg bool, intPart, fracPart string) {
	var sign byte
	if neg {
		sign = '-'
	} else {
		sign = '+'
	}
	sb.WriteByte(sign)
}

func compileTemplate(s string) compiledTemplate {
	sr := strings.NewReader(s)

	ct := compiledTemplate{}

	literal := &strings.Builder{}
	explicitSign := false
	escape := false
	for {
		b, err := sr.ReadByte()
		if err != nil {
			if literal.Len() > 0 {
				ct = append(ct, compiledTemplatePartLiteral(literal.String()))
			}
			break
		}

		if escape {
			literal.WriteByte(b)
			escape = false
			continue
		}

		if b == '\\' {
			escape = true
			continue
		}

		if b == 'n' || b == '-' || b == '+' {
			if literal.Len() > 0 {
				ct = append(ct, compiledTemplatePartLiteral(literal.String()))
				literal.Reset()
			}

			switch b {
			case 'n':
				ct = append(ct, compiledTemplatePartNumber{})
			case '-':
				explicitSign = true
				ct = append(ct, compiledTemplatePartOptionalSign{})
			case '+':
				explicitSign = true
				ct = append(ct, compiledTemplatePartForceSign{})
			}
		} else {
			literal.WriteByte(b)
		}
	}

	if !explicitSign {
		newCt := make(compiledTemplate, 0, len(ct)+1)
		for _, part := range ct {
			if _, ok := part.(compiledTemplatePartNumber); ok {
				ct = append(ct, compiledTemplatePartOptionalSign{})
			}
			newCt = append(newCt, part)
		}
		ct = newCt
	}

	return ct
}

// NewUSDFormatter returns a Formatter for US dollars.
func NewUSDFormatter() *Formatter {
	return &Formatter{
		MinDecimalPlaces: 2,
		Template:         `-$n`,
	}
}

// NewPercentFormatter returns a formatter that formats a number such as 0.75 to 75%.
func NewPercentFormatter() *Formatter {
	return &Formatter{
		Shift:    2,
		Template: `-n%`,
	}
}
