package assertv4

type TestingT interface {
	Helper()
	FailNow()
	Fail()
	Logf(msg string, args ...any)
}

// True fails the test with [t.FailNow] if the value of expr is false.
// True uses the Go source and comments in the test function to print a
// helpful failure message. Extra values can be passed to provide
// the full context for the message.
// If True fails and there are no extra values, it will rewrite the test
// Go source to add any variables that were found in the Go source that
// created expr.
//
// Error is a type (error.As)
//
//	err := ScheduleWork()
//	errV := &CustomErrorType{}
//	assert.True(t, errors.As(err, &errType), err)
//	// main_test.go:17: We wanted err := DoSomeWork() to be &CustomErrorType,
//	// but it was fs.PathError with a value of: file not found.
//
// Error is a value
//
//	fi, err := os.Stat("./output")
//	assert.True(t, errors.Is(err, fs.ErrNotfound), err)
//	// main_test.go:23: We wanted err := os.Stat("./output") to be fs.ErrNotFound,
//	// but it was nil.
//
// Error contains a message
//
//	err := GetAccount()
//	contains := "account id"
//	assert.True(t, err != nil && strings.Contains(err, contains), err)
//	// main_test.go:28: We wanted err := GetAccount() to contain
//	// "account id", but it was "connection failed".
//
// TODO: Error value is message
//
// Map contains a value
//
//	v, ok := settings["max"]
//	assert.True(t, ok, settings)
//	// main_test.go:32: We wanted settings to have the key "max", but it did not.
//	// settings={"min": 34}.
//
// Fail with t.Fail instead of t.FailNow
//
//	assert.True(assert.AndContinue(t),
func True(t TestingT, expr bool, values ...any) bool {
	if expr == true {
		return true
	}
	t.Helper()
	return false
}

// Nil fails the test with [t.FailNow] if the value of err is not nil.
// Nil uses the Go source and comments in the test function to print a
// helpful failure message.
func Nil(t TestingT, err error) bool {
	if err == nil {
		return true
	}
	t.Helper()
	return false
}

func Empty(t TestingT, v string) bool {
	if v == "" {
		return true
	}
	t.Helper()
	return false
}

func Equal(t TestingT, x any, y any) bool {
	if x == y {
		return true
	}
	t.Helper()
	return false
}

// AndContinue returns t with its [t.FailNow] method replaced by [t.Fail].
// Use AndContinue to modify the behaviour of [True], [Nil], [Empty], and [Equal]
// to run the rest of the test, instead of stopping the test immediate.
func AndContinue(t TestingT) TestingT {
	return andContinue{TestingT: t}
}

type andContinue struct {
	TestingT
}

func (c andContinue) FailNow() {
	c.TestingT.Fail()
}
