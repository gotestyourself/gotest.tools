/*Package poll provides tools for testing asynchronous code.
 */
package poll // import "gotest.tools/v3/poll"

import (
	"context"
	"errors"
	"time"
)

// TestingT is the subset of [testing.T] used by [WaitOn]
type TestingT interface {
	LogT
	FailNow()
}

// LogT is a logging interface that is passed to the [WaitOn] check function
type LogT interface {
	Helper()
	Logf(format string, args ...interface{})
}

type delayKeyType struct{}

var delayKey = delayKeyType{}

func WithDelay(ctx context.Context, delay time.Duration) context.Context {
	return context.WithValue(ctx, delayKey, delay)
}

// WaitOn calls check until it returns non-nil error that is not [Continue], or
// the timeout has elapsed. WaitOn sleeps between each call.
// A timeout can be set on the context, and a delay can be set with [WithDelay].
// WaitOn defaults to a 10s timeout and 100ms delay.
//
// The test is failed with [t.FailNow] If WaitOn reaches the timeout, or a non-nil
// error that is not [Continue] is returned by check.
//
// Check should return an error or message using [Continue] to continue waiting,
// and return nil when WaitOn should stop with success.
func WaitOn(ctx context.Context, t TestingT, check Check) {
	t.Helper()

	timeout := 10 * time.Second
	if deadline, hasTimeout := ctx.Deadline(); !hasTimeout {
		var cancel func()
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	} else {
		timeout = time.Until(deadline)
	}

	delay, ok := ctx.Value(delayKey).(time.Duration)
	if !ok {
		delay = 100 * time.Millisecond
	}

	var lastErr error
	chResult := make(chan error)
	for {
		// timeout reached
		if ctx.Err() != nil {
			if lastErr == nil {
				lastErr = errors.New("first check never completed")
			}
			t.Logf("waited %s: %s", timeout, lastErr)
			t.FailNow()
		}

		go func() {
			chResult <- check(ctx, t)
		}()
		select {
		case <-ctx.Done():
			continue
		case err := <-chResult:
			switch {
			case errors.As(err, &cont{}):
				lastErr = err
				select {
				case <-time.After(delay):
					continue
				case <-ctx.Done():
					continue
				}
			case err != nil: // error before timeout
				t.Logf("check failed before timeout: %s", err)
				t.FailNow()
			default:
				return // success
			}
		}
	}
}

type Check func(ctx context.Context, r LogT) error

// Continue wraps an error to indicate to [WaitOn] that it should continue
// waiting.
// The last message returned to [WaitOn] will be used as the failure message
// when [WaitOn] reaches the timeout.
func Continue(err error) error {
	return cont{error: err}
}

type cont struct {
	error
}

func (c cont) Unwrap() error {
	return c.error
}
