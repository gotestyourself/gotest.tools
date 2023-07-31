package poll

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
)

type fakeT struct {
	failed string
}

func (t *fakeT) Helper() {}

func (t *fakeT) Log(args ...interface{}) {}

func (t *fakeT) Logf(format string, args ...interface{}) {
	t.failed = fmt.Sprintf(format, args...)
}

func (t *fakeT) FailNow() {
	panic("exit wait on")

}

func TestWaitOn(t *testing.T) {
	counter := 0
	end := 4
	check := func(ctx context.Context, t LogT) error {
		if counter == end {
			return nil
		}
		counter++
		return Continue(fmt.Errorf("counter is at %d not yet %d", counter-1, end))
	}

	ctx := context.Background()
	WaitOn(ctx, t, check)
	assert.Equal(t, end, counter)
}

func TestWaitOnWithTimeout(t *testing.T) {
	fakeT := &fakeT{}

	check := func(ctx context.Context, t LogT) error {
		return Continue(errors.New("not done"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	assert.Assert(t, cmp.Panics(func() {
		WaitOn(ctx, fakeT, check)
	}))
	assert.Assert(t, strings.Contains(fakeT.failed, "waited"))
	assert.Assert(t, strings.Contains(fakeT.failed, ": not done"))
}

func TestWaitOnWithCheckTimeout(t *testing.T) {
	fakeT := &fakeT{}

	check := func(ctx context.Context, t LogT) error {
		time.Sleep(1 * time.Second)
		return Continue(errors.New("not done"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	assert.Assert(t, cmp.Panics(func() { WaitOn(ctx, fakeT, check) }))
	assert.Assert(t, strings.Contains(fakeT.failed, "waited"))
	assert.Assert(t, strings.Contains(fakeT.failed, ": first check never completed"))
}

func TestWaitOnWithCheckError(t *testing.T) {
	fakeT := &fakeT{}

	check := func(ctx context.Context, t LogT) error {
		return fmt.Errorf("broke")
	}

	ctx := context.Background()
	assert.Assert(t, cmp.Panics(func() { WaitOn(ctx, fakeT, check) }))
	assert.Equal(t, "check failed before timeout: broke", fakeT.failed)
}
