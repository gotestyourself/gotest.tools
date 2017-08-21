package poll

import "time"

var t TestingT

func numOfProcesses() (int, error) {
	return 0, nil
}

func ExampleWaitOn() {
	desired := 10

	check := func(t TestingT) Result {
		actual, err := numOfProcesses()
		if err != nil {
			t.Fatalf("failed to get number of processes: %s", err)
		}
		if actual == desired {
			return Success()
		}
		t.Logf("waiting on process count to be %d...", desired)
		return Continue("number of processes is %d, not %d", actual, desired)
	}

	WaitOn(t, check)
}

func isDesiredState() bool { return false }
func getState() string     { return "" }

func ExampleSettingOp() {
	check := func(t TestingT) Result {
		if isDesiredState() {
			return Success()
		}
		return Continue("state is: %s", getState())
	}
	WaitOn(t, check, WithTimeout(30*time.Second), WithDelay(15*time.Millisecond))
}
