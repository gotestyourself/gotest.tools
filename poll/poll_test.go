package poll

import (
	"fmt"
	"testing"

	"time"

	"github.com/stretchr/testify/assert"
)

type fakeT struct {
	testingT TestingT
	failed   string
}

func (t *fakeT) Fatal(args ...interface{}) {}

func (t *fakeT) Fatalf(format string, args ...interface{}) {
	t.failed = fmt.Sprintf(format, args...)
	panic("exit wait on")
}

func (t *fakeT) Log(args ...interface{}) {}

func (t *fakeT) Logf(format string, args ...interface{}) {}

func TestWaitOn(t *testing.T) {
	counter := 0
	end := 4
	check := func(t TestingT) Result {
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
	fakeT := &fakeT{testingT: t}

	check := func(t TestingT) Result {
		return Continue("not done")
	}

	assert.Panics(t, func() { WaitOn(fakeT, check, WithTimeout(time.Millisecond)) })
	assert.Equal(t, "timeout hit after 1ms: not done", fakeT.failed)
}

func TestWaitOnWithCheckTimeout(t *testing.T) {
	fakeT := &fakeT{testingT: t}

	check := func(t TestingT) Result {
		time.Sleep(1 * time.Second)
		return Continue("not done")
	}

	assert.Panics(t, func() { WaitOn(fakeT, check, WithTimeout(time.Millisecond)) })
	assert.Equal(t, "timeout hit after 1ms: first check never completed", fakeT.failed)
}
