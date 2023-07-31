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

	check := func(ctx context.Context, t poll.LogT) error {
		actual, err := numOfProcesses()
		if err != nil {
			return fmt.Errorf("failed to get number of processes: %w", err)
		}
		if actual == desired {
			return nil
		}
		t.Logf("waiting on process count to be %d...", desired)
		return poll.Continue(fmt.Errorf("number of processes is %d, not %d", actual, desired))
	}

	ctx := context.Background()
	poll.WaitOn(ctx, t, check)
}

func isDesiredState() bool { return false }
func getState() string     { return "" }

func ExampleSettingOp() {
	check := func(ctx context.Context, t poll.LogT) error {
		if isDesiredState() {
			return nil
		}
		return poll.Continue(fmt.Errorf("state is: %s", getState()))
	}

	ctx := poll.WithDelay(context.Background(), 33*time.Millisecond)
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	poll.WaitOn(ctx, t, check)
}
