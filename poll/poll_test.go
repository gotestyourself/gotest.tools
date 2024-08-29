package poll

import (
	"fmt"
	"testing"
	"time"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
)

type fakeT struct {
	failed string
}

func (t *fakeT) Fatalf(format string, args ...interface{}) {
	t.failed = fmt.Sprintf(format, args...)
	panic("exit wait on")
}

func (t *fakeT) Log(...interface{}) {}

func (t *fakeT) Logf(string, ...interface{}) {}

func TestContinueMessage(t *testing.T) {
	tests := []struct {
		msg      string
		args     []interface{}
		expected string
	}{
		{
			msg:      "literal message",
			expected: "literal message",
		},
		{
			msg:      "templated %s",
			args:     []interface{}{"message"},
			expected: "templated message",
		},
		{
			msg:      "literal message with percentage symbols (%USERPROFILE%)",
			expected: "literal message with percentage symbols (%USERPROFILE%)",
		},
	}

	for _, tc := range tests {
		actual := Continue(tc.msg, tc.args...).Message()
		assert.Check(t, cmp.Equal(tc.expected, actual))
	}
}

func TestWaitOn(t *testing.T) {
	counter := 0
	end := 4
	check := func(LogT) Result {
		if counter == end {
			return Success()
		}
		counter++
		return Continue("counter is at %d not yet %d", counter-1, end)
	}

	WaitOn(t, check, WithDelay(0))
	assert.Equal(t, end, counter)
}

func TestWaitOnWithTimeout(t *testing.T) {
	fakeT := &fakeT{}

	check := func(LogT) Result {
		return Continue("not done")
	}

	assert.Assert(t, cmp.Panics(func() {
		WaitOn(fakeT, check, WithTimeout(time.Millisecond))
	}))
	assert.Equal(t, "timeout hit after 1ms: not done", fakeT.failed)
}

func TestWaitOnWithCheckTimeout(t *testing.T) {
	fakeT := &fakeT{}

	check := func(LogT) Result {
		time.Sleep(1 * time.Second)
		return Continue("not done")
	}

	assert.Assert(t, cmp.Panics(func() { WaitOn(fakeT, check, WithTimeout(time.Millisecond)) }))
	assert.Equal(t, "timeout hit after 1ms: first check never completed", fakeT.failed)
}

func TestWaitOnWithCheckError(t *testing.T) {
	fakeT := &fakeT{}

	check := func(LogT) Result {
		return Error(fmt.Errorf("broke"))
	}

	assert.Assert(t, cmp.Panics(func() { WaitOn(fakeT, check) }))
	assert.Equal(t, "polling check failed: broke", fakeT.failed)
}

func TestWaitOn_WithCompare(t *testing.T) {
	fakeT := &fakeT{}

	check := func(LogT) Result {
		return Compare(cmp.Equal(3, 4))
	}

	assert.Assert(t, cmp.Panics(func() {
		WaitOn(fakeT, check, WithDelay(0), WithTimeout(10*time.Millisecond))
	}))
	assert.Assert(t, cmp.Contains(fakeT.failed, "assertion failed: 3 (int) != 4 (int)"))
}
