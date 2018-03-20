package format_test

import (
	"testing"

	"github.com/gotestyourself/gotestyourself/assert"
	"github.com/gotestyourself/gotestyourself/golden"
	"github.com/gotestyourself/gotestyourself/internal/format"
)

func TestUnifiedDiff(t *testing.T) {
	var testcases = []struct {
		name         string
		a            string
		b            string
		expected     string
		expectedFile string
		from         string
		to           string
	}{
		{
			name: "empty diff",
			a:    "a\nb\nc",
			b:    "a\nb\nc",
			from: "from",
			to:   "to",
		},
		{
			name:         "one diff with header",
			a:            "a\nxyz\nc",
			b:            "a\nb\nc",
			from:         "from",
			to:           "to",
			expectedFile: "one-diff-with-header.golden",
		},
		{
			name:         "many diffs",
			a:            "a123\nxyz\nc\nbaba\nz\nt\nj2j2\nok\nok\ndone\n",
			b:            "a123\nxyz\nc\nabab\nz\nt\nj2j2\nok\nok\n",
			expectedFile: "many-diff.golden",
		},
		{
			name:         "no trailing newline",
			a:            "a123\nxyz\nc\nbaba\nz\nt\nj2j2\nok\nok\ndone\n",
			b:            "a123\nxyz\nc\nabab\nz\nt\nj2j2\nok\nok",
			expectedFile: "many-diff-no-trailing-newline.golden",
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			diff := format.UnifiedDiff(format.DiffConfig{
				A:    testcase.a,
				B:    testcase.b,
				From: testcase.from,
				To:   testcase.to,
			})

			if testcase.expectedFile != "" {
				assert.Assert(t, golden.String(diff, testcase.expectedFile))
				return
			}
			assert.Equal(t, diff, testcase.expected)
		})
	}
}
