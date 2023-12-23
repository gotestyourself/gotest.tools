package poll_test

import (
	"context"
	"fmt"
	"time"

	"gotest.tools/v3/poll"
)

var t poll.TestingT

func numOfProcesses() (int, error) {
	return 0, nil
}

func ExampleWaitOn() {
	desired := 10

	ctx := context.Background()
	poll.WaitOn(ctx, t, func(ctx context.Context) error {
		actual, err := numOfProcesses()
		if err != nil {
			return fmt.Errorf("failed to get number of processes: %w", err)
		}
		if actual == desired {
			return nil
		}
		t.Logf("waiting on process count to be %d...", desired)
		return poll.Continue(fmt.Errorf("number of processes is %d, not %d", actual, desired))
	})
}

func isDesiredState() bool { return false }
func getState() string     { return "" }

func ExampleWaitOn_WithDelay() {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	poll.WaitOn(poll.WithDelay(ctx, 33*time.Millisecond), t, func(ctx context.Context) error {
		if isDesiredState() {
			return nil
		}
		return poll.Continue(fmt.Errorf("state is: %s", getState()))
	})
}
