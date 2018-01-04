/*Package assert provides assertions and checks for comparing expected values to
actual values. When an assertion or check fails a helpful error message is
printed.

Assert and Check

Assert() and Check() both accept a Comparison, and fail the test when the
comparison fails. The one difference is that Assert() will end the test execution
immediately (using t.FailNow()) whereas Check() will fail the test (using t.Fail()),
return the value of the comparison, then proceed with the rest of the test case.

Example Usage

The example below shows assert used with some common types.


	import (
	    "testing"

	    "github.com/gotestyourself/gotestyourself/assert"
	    is "github.com/gotestyourself/gotestyourself/assert/cmp"
	)

	func TestEverything(t *testing.T) {
	    // booleans
	    assert.Assert(t, isOk)
	    assert.Assert(t, !missing)

	    // primitives
	    assert.Equal(t, count, 1)
	    assert.Equal(t, msg, "the message")
	    assert.Assert(t, total != 10) // NotEqual

	    // errors
	    assert.NilError(t, closer.Close())
	    assert.Assert(t, is.Error(err, "the exact error message"))
	    assert.Assert(t, is.ErrorContains(err, "includes this"))

	    // complex types
	    assert.Assert(t, is.Len(items, 3))
	    assert.Assert(t, len(sequence) != 0) // NotEmpty
	    assert.Assert(t, is.Contains(mapping, "key"))
	    assert.Assert(t, is.Compare(result, myStruct{Name: "title"}))

	    // pointers and interface
	    assert.Assert(t, is.Nil(ref))
	    assert.Assert(t, ref != nil) // NotNil
	}

Comparisons

https://godoc.org/github.com/gotestyourself/gotestyourself/assert/cmp provides
many common comparisons. Additional comparisons can be written to compare
values in other ways.

Below is an example of a custom comparison using a regex pattern:

	func RegexP(value string, pattern string) func() (bool, string) {
	    return func() (bool, string) {
	        re := regexp.MustCompile(pattern)
	        msg := fmt.Sprintf("%q did not match pattern %q", value, pattern)
	        return re.MatchString(value), msg
	    }
	}

*/
package assert

import (
	"fmt"

	"github.com/gotestyourself/gotestyourself/assert/cmp"
	"github.com/gotestyourself/gotestyourself/internal/format"
	"github.com/gotestyourself/gotestyourself/internal/source"
)

// BoolOrComparison can be a bool, or Comparison, other types will panic.
type BoolOrComparison interface{}

// Comparison is a function which compares values and returns true if the actual
// value matches the expected value. If the values do not match it returns a message
// with details about why it failed.
//
// https://godoc.org/github.com/gotestyourself/gotestyourself/assert/cmp
// provides many general purpose Comparisons.
type Comparison func() (success bool, message string)

// TestingT is the subset of testing.T used by the assert package
type TestingT interface {
	FailNow()
	Fail()
	Log(args ...interface{})
}

type helperT interface {
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

// New returns a new Tester for asserting and checking values.
func New(t TestingT) Tester {
	return Tester{t: t, stackIndex: stackIndex, argPos: 0}
}

// Assert performs a comparison, marks the test as having failed if the comparison
// returns false, and stops execution immediately.
func (t Tester) Assert(comparison BoolOrComparison, msgAndArgs ...interface{}) {
	if ht, ok := t.t.(helperT); ok {
		ht.Helper()
	}
	t.assert(t.t.FailNow, comparison, msgAndArgs...)
}

func (t Tester) assert(failer func(), comparison BoolOrComparison, msgAndArgs ...interface{}) bool {
	if ht, ok := t.t.(helperT); ok {
		ht.Helper()
	}
	switch check := comparison.(type) {
	case bool:
		if check {
			return true
		}
		source, err := source.GetCondition(t.stackIndex, t.argPos)
		if err != nil {
			t.t.Log(err.Error())
		}

		msg := " is false"
		t.t.Log(format.WithCustomMessage(failureMessage+source+msg, msgAndArgs...))
		failer()
		return false

	case Comparison:
		return runCompareFunc(failer, t.t, check, msgAndArgs...)

	case func() (success bool, message string):
		return runCompareFunc(failer, t.t, check, msgAndArgs...)

	default:
		panic(fmt.Sprintf("invalid type for condition arg: %T", comparison))
	}
}

func runCompareFunc(failer func(), t TestingT, f Comparison, msgAndArgs ...interface{}) bool {
	if ht, ok := t.(helperT); ok {
		ht.Helper()
	}
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
	if ht, ok := t.t.(helperT); ok {
		ht.Helper()
	}
	return t.assert(t.t.Fail, comparison, msgAndArgs...)
}

// NilError fails the test immediately if the last arg is a non-nil error.
// This is equivalent to Assert(cmp.NilError(err))
func (t Tester) NilError(arg interface{}, args ...interface{}) {
	if ht, ok := t.t.(helperT); ok {
		ht.Helper()
	}
	t.assert(t.t.FailNow, cmp.NilError(arg, args...))
}

// Equal uses the == operator to assert two values are equal and fails the test
// if they are not equal.
// This is equivalent to Assert(cmp.Equal(x, y))
func (t Tester) Equal(x, y interface{}, msgAndArgs ...interface{}) {
	if ht, ok := t.t.(helperT); ok {
		ht.Helper()
	}
	t.assert(t.t.FailNow, cmp.Equal(x, y), msgAndArgs...)
}

// Assert performs a comparison and fails the test immediate if the comparison
// is not a success.
func Assert(t TestingT, comparison BoolOrComparison, msgAndArgs ...interface{}) {
	if ht, ok := t.(helperT); ok {
		ht.Helper()
	}
	newPackageScopeTester(t).Assert(comparison, msgAndArgs...)
}

// Check performs a comparison and marks the test as having failed if the comparison
// returns false. Returns the result of the comparison.
func Check(t TestingT, comparison BoolOrComparison, msgAndArgs ...interface{}) bool {
	if ht, ok := t.(helperT); ok {
		ht.Helper()
	}
	return newPackageScopeTester(t).Check(comparison, msgAndArgs...)
}

// NilError fails the test immediately if the last arg is a non-nil error.
func NilError(t TestingT, arg interface{}, args ...interface{}) {
	if ht, ok := t.(helperT); ok {
		ht.Helper()
	}
	newPackageScopeTester(t).NilError(arg, args...)
}

// Equal uses the == operator to assert two values are equal, and fails the test
// if they are not equal.
func Equal(t TestingT, x, y interface{}, msgAndArgs ...interface{}) {
	if ht, ok := t.(helperT); ok {
		ht.Helper()
	}
	newPackageScopeTester(t).Equal(x, y, msgAndArgs...)
}

// newPackageScopeTester returns a Tester appropriate for package level functions.
// The tester has stackIndex+1 to accommodate the extra function in the stack, and
// argPos 1 because package level functions accept testing.T as the first argument
func newPackageScopeTester(t TestingT) Tester {
	return Tester{t: t, stackIndex: stackIndex + 1, argPos: 1}
}
