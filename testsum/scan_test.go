package testsum

import (
	"bytes"
	"strings"
	"testing"

	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScanNoFailures(t *testing.T) {
	source := `=== RUN   TestRunCommandSuccess
--- PASS: TestRunCommandSuccess (0.00s)
=== RUN   TestRunCommandWithCombined
--- PASS: TestRunCommandWithCombined (0.00s)
=== RUN   TestRunCommandWithTimeoutFinished
--- PASS: TestRunCommandWithTimeoutFinished (0.00s)
=== RUN   TestRunCommandWithTimeoutKilled
--- PASS: TestRunCommandWithTimeoutKilled (1.25s)
=== RUN   TestRunCommandWithErrors
--- PASS: TestRunCommandWithErrors (0.00s)
=== RUN   TestRunCommandWithStdoutStderr
--- PASS: TestRunCommandWithStdoutStderr (0.00s)
=== RUN   TestRunCommandWithStdoutStderrError
--- PASS: TestRunCommandWithStdoutStderrError (0.00s)
=== RUN   TestSkippedBecauseSomething
--- SKIP: TestSkippedBecauseSomething (0.00s)
        scan_test.go:39: becausde blah
PASS
ok      github.com/gotestyourself/gotestyourself/icmd   1.256s
`

	out := new(bytes.Buffer)
	summary, err := Scan(strings.NewReader(source), out)
	require.NoError(t, err)
	assert.NotZero(t, summary.Elapsed)
	summary.Elapsed = 0 // ignore elapsed
	assert.Equal(t, &Summary{Total: 8, Skipped: 1}, summary)
	assert.Equal(t, source, out.String())
}

func TestScanWithFailure(t *testing.T) {
	source := `=== RUN   TestRunCommandWithCombined
--- PASS: TestRunCommandWithCombined (0.00s)
=== RUN   TestRunCommandWithStdoutStderrError
--- PASS: TestRunCommandWithStdoutStderrError (0.00s)
=== RUN   TestThisShouldFail
Some output
More output
--- FAIL: TestThisShouldFail (0.00s)
        dummy_test.go:11: test is bad
        dummy_test.go:12: another failure
FAIL
exit status 1
FAIL    github.com/gotestyourself/gotestyourself/testsum        0.002s
`

	out := new(bytes.Buffer)
	summary, err := Scan(strings.NewReader(source), out)
	require.NoError(t, err)
	assert.NotZero(t, summary.Elapsed)
	summary.Elapsed = 0 // ignore elapsed
	assert.Equal(t, source, out.String())

	expected := &Summary{
		Total: 3,
		Failures: []Failure{
			{
				name:   "TestThisShouldFail",
				output: "Some output\nMore output\n",
				logs:   "        dummy_test.go:11: test is bad\n        dummy_test.go:12: another failure\n",
			},
		},
	}
	assert.Equal(t, expected, summary)
}

func TestSummaryFormatLine(t *testing.T) {
	var testcases = []struct {
		summary  Summary
		expected string
	}{
		{
			summary:  Summary{Total: 15, Elapsed: time.Minute},
			expected: "======== 15 tests in 60.00 seconds ========",
		},
		{
			summary:  Summary{Total: 100, Skipped: 3},
			expected: "======== 100 tests, 3 skipped in 0.00 seconds ========",
		},
		{
			summary: Summary{
				Total:    100,
				Failures: []Failure{{}},
				Elapsed:  3555 * time.Millisecond,
			},
			expected: "======== 100 tests, 1 failed in 3.56 seconds ========",
		},
		{
			summary: Summary{
				Total:    100,
				Skipped:  3,
				Failures: []Failure{{}},
				Elapsed:  42,
			},
			expected: "======== 100 tests, 3 skipped, 1 failed in 0.00 seconds ========",
		},
	}

	for _, testcase := range testcases {
		assert.Equal(t, testcase.expected, testcase.summary.FormatLine())
	}
}
