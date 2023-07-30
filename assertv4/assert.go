package assertv4

type TestingT interface {
	Helper()
	FailNow()
	Logf(msg string, args ...any)
}

// True fails the test with [t.FailNow] if the value of expr is false.
// True uses the Go source and comments in the test function to print a
// helpful failure message. Extra values can be passed to provide
// the full context for the message.
// If True sees the test fail and there are
// no extra values, it will rewrite the test Go source to add any variables
// that were found in the Go source that created expr.
//
//		err := ScheduleWork()
//		errV := &CustomErrorType{}
//		assert.True(t, errors.As(err, &errType), err)
//		// main_test.go:17: We wanted err := DoSomeWork() to be &CustomErrorType,
//		// but it was fs.PathError with a value of: file not found.
//
//		fi, err := os.Stat("./output")
//		assert.True(t, errors.Is(err, fs.ErrNotfound), err)
//		// main_test.go:23: We wanted err := os.Stat("./output") to be fs.ErrNotFound,
//		// but it was nil.
//
//	 err := GetAccount()
//	 contains := "account id"
//	 assert.True(t, err != nil && strings.Contains(err, contains), err)
//	 // main_Test.go:28: We wanted err := GetAccount() to contain
//	 // "account id", but it was "connection failed".
func True(t TestingT, expr bool, values ...any) {
	if expr == true {
		return
	}
	t.Helper()

}

// Nil fails the test with [t.FailNow] if the value of err is not nil.
// Nil uses the Go source and comments in the test function to print a
// helpful failure message.
func Nil(t TestingT, err error) {
	if err == nil {
		return
	}
	t.Helper()

}

func Empty(t TestingT, v string) {
	if v == "" {
		return
	}
	t.Helper()

}

func Equal(t TestingT, x any, y any) {
	if x == y {
		return
	}
	t.Helper()

}
