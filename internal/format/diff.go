package format

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/gotestyourself/gotestyourself/internal/difflib"
)

const (
	contextLines = 2
)

// DiffConfig for a unified diff
type DiffConfig struct {
	A    string
	B    string
	From string
	To   string
}

// TODO: handle diff where prefix or suffix is only whitespace

// UnifiedDiff is a modified version of difflib.WriteUnifiedDiff with better
// support for showing the whitespace differences.
func UnifiedDiff(conf DiffConfig) string {
	a := strings.SplitAfter(conf.A, "\n")
	b := strings.SplitAfter(conf.B, "\n")
	groups := difflib.NewMatcher(a, b).GetGroupedOpCodes(contextLines)
	if len(groups) == 0 {
		return ""
	}

	buf := new(bytes.Buffer)
	wf := func(format string, args ...interface{}) {
		buf.WriteString(fmt.Sprintf(format, args...))
	}
	formatHeader(wf, conf)
	for _, g := range groups {
		formatRangeLine(wf, g)
		for _, c := range g {
			in, out := a[c.I1:c.I2], b[c.J1:c.J2]
			switch c.Tag {
			case 'e':
				formatLines(buf, " ", in)
			case 'r':
				formatLines(buf, "-", in)
				formatLines(buf, "+", out)
			case 'd':
				formatLines(buf, "-", in)
			case 'i':
				formatLines(buf, "+", out)
			}
		}
	}
	return buf.String()
}

func formatHeader(wf func(string, ...interface{}), conf DiffConfig) {
	if conf.From != "" || conf.To != "" {
		wf("--- %s\n", conf.From)
		wf("+++ %s\n", conf.To)
	}
}

func formatRangeLine(wf func(string, ...interface{}), group []difflib.OpCode) {
	first, last := group[0], group[len(group)-1]
	range1 := formatRangeUnified(first.I1, last.I2)
	range2 := formatRangeUnified(first.J1, last.J2)
	wf("@@ -%s +%s @@\n", range1, range2)
}

// Convert range to the "ed" format
func formatRangeUnified(start, stop int) string {
	// Per the diff spec at http://www.unix.org/single_unix_specification/
	beginning := start + 1 // lines start numbering with one
	length := stop - start
	if length == 1 {
		return fmt.Sprintf("%d", beginning)
	}
	if length == 0 {
		beginning-- // empty ranges begin at line just before the range
	}
	return fmt.Sprintf("%d,%d", beginning, length)
}

func formatLines(buf *bytes.Buffer, prefix string, lines []string) {
	for _, line := range lines {
		buf.WriteString(prefix + line)
	}
	// Add a newline if the last line is missing one so that the diff displays
	// properly.
	if !strings.HasSuffix(lines[len(lines)-1], "\n") {
		buf.WriteString("\n")
	}
}
