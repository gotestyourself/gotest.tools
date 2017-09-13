package tags

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type fakeSkipT struct {
	message string
}

func (f *fakeSkipT) Skipf(format string, args ...interface{}) {
	f.message = fmt.Sprintf(format, args...)
	panic("exit apply")
}

func TestApply(t *testing.T) {
	oldRequested := requested[:]
	defer func() { requested = oldRequested }()

	var testcases = []struct {
		doc      string
		tags     []string
		applied  []string
		expected string
	}{
		{
			doc:     "no tags",
			applied: list("any", "none"),
		},
		{
			doc:      "no applied tags does not match any",
			tags:     list("something"),
			expected: "no matching tag",
		},
		{
			doc:     "short any tags",
			tags:    list("a", "b", "c"),
			applied: list("b"),
		},
		{
			doc:      "does not match any tag",
			tags:     list("one", "two"),
			applied:  list("not", "this"),
			expected: "no matching tag",
		},
		{
			doc:      "matches skip tag",
			tags:     list("!one"),
			applied:  list("one", "two"),
			expected: "matched tag: !one",
		},
		{
			doc:     "no matching skip tag",
			tags:    list("!one"),
			applied: list("two", "three"),
		},

		{
			doc:     "all required tags match",
			tags:    list("+one", "+two", "+three"),
			applied: list("one", "two", "three", "four"),
		},
		{
			doc:      "all required tags do not matched",
			tags:     list("+one", "+two"),
			applied:  list("one"),
			expected: "missing required tag: two",
		},
		{
			doc:      "matches all required and a skipped",
			tags:     list("+yes", "+do", "!skipme"),
			applied:  list("yes", "do", "skipme"),
			expected: "matched tag: !skipme",
		},
		{
			doc:     "matches all required and no skipped",
			tags:    list("+yes", "+do", "!skipme"),
			applied: list("yes", "do"),
		},
		{
			doc:      "matches an any but not all required",
			tags:     list("sure", "+do", "+reqone"),
			applied:  list("sure", "do"),
			expected: "missing required tag: reqone",
		},
		{
			doc:      "matches an any and a skipped",
			tags:     list("sure", "!skipme"),
			applied:  list("sure", "skipme"),
			expected: "matched tag: !skipme",
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.doc, func(t *testing.T) {
			fakeT := &fakeSkipT{}
			requested = testcase.tags
			handlePanic(func() { Apply(fakeT, testcase.applied...) })
			assert.Equal(t, testcase.expected, fakeT.message)
		})
	}
}

func list(arg ...string) []string {
	return arg
}

func handlePanic(f func()) {
	defer func() { _ = recover() }()
	f()
}
