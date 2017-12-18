/*Package assert provides assertions and checks for comparing expected values to
actual values, and printing helpful failure messages.
*/
package assert

import (
	"fmt"

	"github.com/gotestyourself/gotestyourself/assert/cmp"
	"github.com/gotestyourself/gotestyourself/internal/format"
	"github.com/gotestyourself/gotestyourself/internal/source"
)

// BoolOrComparison can be a bool, Comparison, or CompareFunc, other types will
// panic
type BoolOrComparison interface{}

// Comparison provides a compare method for comparing values
type Comparison interface {
	// Compare performs a comparison and returns true if actual value matches
	// the expected value. If the values do not match it returns a message
	// with details about why it failNowed.
	Compare() (success bool, message string)
}

// CompareFunc is a Comparison.Compare()
type CompareFunc func() (success bool, message string)

// TestingT is the subset of testing.T used by the assert package
type TestingT interface {
	FailNow()
	Fail()
	Log(args ...interface{})
	Helper()
}

// Tester wraps a TestingT and provides assertions and checks
type Tester struct {
	t          TestingT
	stackIndex int
	argPos     int
}

// stackIndex = Assert()/Check(), assert()
const stackIndex = 2

const failureMessage = "assertion failed: "

// New returns a new Tester for asserting and checking values
func New(t TestingT) Tester {
	return Tester{t: t, stackIndex: stackIndex, argPos: 0}
}

// Assert performs a comparison, marks the test as having failed if the comparison
// returns false, and stops execution immediately.
func (t Tester) Assert(comparison BoolOrComparison, msgAndArgs ...interface{}) {
	t.t.Helper()
	t.assert(t.t.FailNow, comparison, msgAndArgs...)
}

func (t Tester) assert(failer func(), comparison BoolOrComparison, msgAndArgs ...interface{}) bool {
	t.t.Helper()
	switch check := comparison.(type) {
	case bool:
		if check {
			return true
		}
		source, err := source.GetCondition(t.stackIndex, t.argPos)
		if err != nil {
			t.t.Log(err.Error())
		}

		t.t.Log(format.WithCustomMessage(failureMessage+source, msgAndArgs...))
		failer()
		return false

	case Comparison:
		return runCompareFunc(failer, t.t, check.Compare, msgAndArgs...)

	case func() (success bool, message string):
		return runCompareFunc(failer, t.t, check, msgAndArgs...)

	case CompareFunc:
		return runCompareFunc(failer, t.t, check, msgAndArgs...)

	default:
		panic(fmt.Sprintf("invalid type for condition arg: %T", comparison))
	}
}

func runCompareFunc(failer func(), t TestingT, f CompareFunc, msgAndArgs ...interface{}) bool {
	t.Helper()
	if success, message := f(); !success {
		t.Log(format.WithCustomMessage(failureMessage+message, msgAndArgs...))
		failer()
		return false
	}
	return true
}

// Check performs a comparison and marks the test as having failed if the comparison
// returns false. Returns the result of the comparison.
func (t Tester) Check(comparison BoolOrComparison, msgAndArgs ...interface{}) bool {
	t.t.Helper()
	return t.assert(t.t.Fail, comparison, msgAndArgs...)
}

// NoError fails the test immediately if the last arg is a non-nil error
func (t Tester) NoError(args ...interface{}) {
	t.t.Helper()
	if len(args) == 0 {
		return
	}
	switch lastArg := args[len(args)-1].(type) {
	case error:
		t.t.Log(fmt.Sprintf("expected no error, got %s", lastArg))
		t.t.FailNow()
	case nil:
	default:
		t.t.Log(fmt.Sprintf("last argument to NoError() must be an error, got %T", lastArg))
		t.t.FailNow()
	}
}

// Equal uses the == operator to assert two values are the equal
func (t Tester) Equal(x, y interface{}, msgAndArgs ...interface{}) {
	t.t.Helper()
	t.assert(t.t.FailNow, cmp.Equal(x, y), msgAndArgs...)
}

// Assert fails the test immediate if comparison is not a success
func Assert(t TestingT, comparison BoolOrComparison, msgAndArgs ...interface{}) {
	t.Helper()
	newPackageScopeTester(t).Assert(comparison, msgAndArgs...)
}

// Check performs a comparison and marks the test as having failed if the comparison
// returns false. Returns the result of the comparison.
func Check(t TestingT, comparison BoolOrComparison, msgAndArgs ...interface{}) bool {
	t.Helper()
	return newPackageScopeTester(t).Check(comparison, msgAndArgs...)
}

// NoError fails the test immediately if the last arg is a non-nil error
func NoError(t TestingT, args ...interface{}) {
	t.Helper()
	newPackageScopeTester(t).NoError(args...)
}

// Equal uses the == operator to assert two values are the equal
func Equal(t TestingT, x, y interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	newPackageScopeTester(t).Equal(x, y, msgAndArgs...)
}

// newPackageScopeTester returns a Tester appropriate for package level functions.
// The tester has stackIndex+1 to accommodate the extra function in the stack, and
// argPos 1 because package level functions accept testing.T as the first argument
func newPackageScopeTester(t TestingT) Tester {
	return Tester{t: t, stackIndex: stackIndex + 1, argPos: 1}
}
