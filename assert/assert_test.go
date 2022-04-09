package assert_test

import (
	"fmt"
	"os"
	"testing"

	gocmp "github.com/google/go-cmp/cmp"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
)

type fakeTestingT struct {
	failNowed bool
	failed    bool
	msgs      []string
}

func (f *fakeTestingT) FailNow() {
	f.failNowed = true
}

func (f *fakeTestingT) Fail() {
	f.failed = true
}

func (f *fakeTestingT) Log(args ...interface{}) {
	f.msgs = append(f.msgs, args[0].(string))
}

func (f *fakeTestingT) Helper() {}

func TestAssertWithBoolFailure(t *testing.T) {
	fakeT := &fakeTestingT{}

	assert.Assert(fakeT, 1 == 6)
	expectFailNowed(t, fakeT, "assertion failed: expression is false: 1 == 6")
}

func TestAssertWithBoolFailureNotEqual(t *testing.T) {
	fakeT := &fakeTestingT{}

	var err error
	assert.Assert(fakeT, err != nil)
	expectFailNowed(t, fakeT, "assertion failed: err is nil")
}

func TestAssertWithBoolFailureNotTrue(t *testing.T) {
	fakeT := &fakeTestingT{}

	badNews := true
	assert.Assert(fakeT, !badNews)
	expectFailNowed(t, fakeT, "assertion failed: badNews is true")
}

func TestAssertWithBoolFailureAndExtraMessage(t *testing.T) {
	fakeT := &fakeTestingT{}

	assert.Assert(fakeT, 1 > 5, "sometimes things fail")
	expectFailNowed(t, fakeT,
		"assertion failed: expression is false: 1 > 5: sometimes things fail")
}

func TestAssertWithBoolSuccess(t *testing.T) {
	fakeT := &fakeTestingT{}

	assert.Assert(fakeT, 1 < 5)
	expectSuccess(t, fakeT)
}

func TestAssertWithBoolMultiLineFailure(t *testing.T) {
	fakeT := &fakeTestingT{}

	assert.Assert(fakeT, func() bool {
		for range []int{1, 2, 3, 4} {
		}
		return false
	}())
	expectFailNowed(t, fakeT, `assertion failed: expression is false: func() bool {
	for range []int{1, 2, 3, 4} {
	}
	return false
}()`)
}

type exampleComparison struct {
	success bool
	message string
}

func (c exampleComparison) Compare() (bool, string) {
	return c.success, c.message
}

func TestAssertWithComparisonSuccess(t *testing.T) {
	fakeT := &fakeTestingT{}

	cmp := exampleComparison{success: true}
	assert.Assert(fakeT, cmp.Compare)
	expectSuccess(t, fakeT)
}

func TestAssertWithComparisonFailure(t *testing.T) {
	fakeT := &fakeTestingT{}

	cmp := exampleComparison{message: "oops, not good"}
	assert.Assert(fakeT, cmp.Compare)
	expectFailNowed(t, fakeT, "assertion failed: oops, not good")
}

func TestAssertWithComparisonAndExtraMessage(t *testing.T) {
	fakeT := &fakeTestingT{}

	cmp := exampleComparison{message: "oops, not good"}
	assert.Assert(fakeT, cmp.Compare, "extra stuff %v", true)
	expectFailNowed(t, fakeT, "assertion failed: oops, not good: extra stuff true")
}

type customError struct {
	field bool
}

func (e *customError) Error() string {
	// access a field of the receiver to simulate the behaviour of most
	// implementations, and test handling of non-nil typed errors.
	e.field = true
	return "custom error"
}

func TestNilError(t *testing.T) {
	t.Run("nil interface", func(t *testing.T) {
		fakeT := &fakeTestingT{}
		var err error
		assert.NilError(fakeT, err)
		expectSuccess(t, fakeT)
	})

	t.Run("nil literal", func(t *testing.T) {
		fakeT := &fakeTestingT{}
		assert.NilError(fakeT, nil)
		expectSuccess(t, fakeT)
	})

	t.Run("interface with non-nil type", func(t *testing.T) {
		fakeT := &fakeTestingT{}
		var customErr *customError
		assert.NilError(fakeT, customErr)
		expected := "assertion failed: error is not nil: error has type *assert_test.customError"
		expectFailNowed(t, fakeT, expected)
	})

	t.Run("non-nil error", func(t *testing.T) {
		fakeT := &fakeTestingT{}
		assert.NilError(fakeT, fmt.Errorf("this is the error"))
		expectFailNowed(t, fakeT, "assertion failed: error is not nil: this is the error")
	})

	t.Run("non-nil error with struct type", func(t *testing.T) {
		fakeT := &fakeTestingT{}
		err := structError{}
		assert.NilError(fakeT, err)
		expectFailNowed(t, fakeT, "assertion failed: error is not nil: this is a struct")
	})

	t.Run("non-nil error with map type", func(t *testing.T) {
		fakeT := &fakeTestingT{}
		var err mapError
		assert.NilError(fakeT, err)
		expectFailNowed(t, fakeT, "assertion failed: error is not nil: ")
	})
}

type structError struct{}

func (structError) Error() string {
	return "this is a struct"
}

type mapError map[int]string

func (m mapError) Error() string {
	return m[0]
}

func TestCheckFailure(t *testing.T) {
	fakeT := &fakeTestingT{}

	if assert.Check(fakeT, 1 == 2) {
		t.Error("expected check to return false on failure")
	}
	expectFailed(t, fakeT, "assertion failed: expression is false: 1 == 2")
}

func TestCheckSuccess(t *testing.T) {
	fakeT := &fakeTestingT{}

	if !assert.Check(fakeT, true) {
		t.Error("expected check to return true on success")
	}
	expectSuccess(t, fakeT)
}

func TestCheckEqualFailure(t *testing.T) {
	fakeT := &fakeTestingT{}

	actual, expected := 5, 9
	assert.Check(fakeT, cmp.Equal(actual, expected))
	expectFailed(t, fakeT, "assertion failed: 5 (actual int) != 9 (expected int)")
}

func TestCheck_MultipleFunctionsOnTheSameLine(t *testing.T) {
	fakeT := &fakeTestingT{}

	f := func(b bool) {}
	f(assert.Check(fakeT, false))
	expectFailed(t, fakeT, "assertion failed: expression is false: false")
}

func TestEqualSuccess(t *testing.T) {
	fakeT := &fakeTestingT{}

	assert.Equal(fakeT, 1, 1)
	expectSuccess(t, fakeT)

	assert.Equal(fakeT, "abcd", "abcd")
	expectSuccess(t, fakeT)
}

func TestEqualFailure(t *testing.T) {
	fakeT := &fakeTestingT{}

	actual, expected := 1, 3
	assert.Equal(fakeT, actual, expected)
	expectFailNowed(t, fakeT, "assertion failed: 1 (actual int) != 3 (expected int)")
}

func TestEqualFailureTypes(t *testing.T) {
	fakeT := &fakeTestingT{}

	assert.Equal(fakeT, 3, uint(3))
	expectFailNowed(t, fakeT, `assertion failed: 3 (int) != 3 (uint)`)
}

func TestEqualFailureWithSelectorArgument(t *testing.T) {
	fakeT := &fakeTestingT{}

	type tc struct {
		expected string
	}
	var testcase = tc{expected: "foo"}

	assert.Equal(fakeT, "ok", testcase.expected)
	expectFailNowed(t, fakeT,
		"assertion failed: ok (string) != foo (testcase.expected string)")
}

func TestEqualFailureWithIndexExpr(t *testing.T) {
	fakeT := &fakeTestingT{}

	expected := map[string]string{"foo": "bar"}
	assert.Equal(fakeT, "ok", expected["foo"])
	expectFailNowed(t, fakeT,
		`assertion failed: ok (string) != bar (expected["foo"] string)`)
}

func TestEqualFailureWithCallExprArgument(t *testing.T) {
	fakeT := &fakeTestingT{}
	ce := customError{}
	assert.Equal(fakeT, "", ce.Error())
	expectFailNowed(t, fakeT,
		"assertion failed:  (string) != custom error (string)")
}

func TestAssertFailureWithOfflineComparison(t *testing.T) {
	fakeT := &fakeTestingT{}
	a := 1
	b := 2
	// store comparison in a variable, so ast lookup can't find it
	comparison := cmp.Equal(a, b)
	assert.Assert(fakeT, comparison)
	// expected value wont have variable names
	expectFailNowed(t, fakeT, "assertion failed: 1 (int) != 2 (int)")
}

type testingT interface {
	Errorf(msg string, args ...interface{})
	Fatalf(msg string, args ...interface{})
	Helper()
}

func expectFailNowed(t testingT, fakeT *fakeTestingT, expected string) {
	t.Helper()
	if fakeT.failed {
		t.Errorf("should not have failed, got messages %s", fakeT.msgs)
	}
	if !fakeT.failNowed {
		t.Fatalf("should have failNowed with message %s", expected)
	}
	if fakeT.msgs[0] != expected {
		t.Fatalf("should have failure message %q, got %q", expected, fakeT.msgs[0])
	}
}

// nolint: unparam
func expectFailed(t testingT, fakeT *fakeTestingT, expected string) {
	t.Helper()
	if fakeT.failNowed {
		t.Errorf("should not have failNowed, got messages %s", fakeT.msgs)
	}
	if !fakeT.failed {
		t.Fatalf("should have failed with message %s", expected)
	}
	if fakeT.msgs[0] != expected {
		t.Fatalf("should have failure message %q, got %q", expected, fakeT.msgs[0])
	}
}

func expectSuccess(t testingT, fakeT *fakeTestingT) {
	t.Helper()
	if fakeT.failNowed {
		t.Errorf("should not have failNowed, got messages %s", fakeT.msgs)
	}
	if fakeT.failed {
		t.Errorf("should not have failed, got messages %s", fakeT.msgs)
	}
}

type stub struct {
	a string
	b int
}

func TestDeepEqualSuccess(t *testing.T) {
	actual := stub{"ok", 1}
	expected := stub{"ok", 1}

	fakeT := &fakeTestingT{}
	assert.DeepEqual(fakeT, actual, expected, gocmp.AllowUnexported(stub{}))
	expectSuccess(t, fakeT)
}

func TestDeepEqualFailure(t *testing.T) {
	actual := stub{"ok", 1}
	expected := stub{"ok", 2}

	fakeT := &fakeTestingT{}
	assert.DeepEqual(fakeT, actual, expected, gocmp.AllowUnexported(stub{}))
	if !fakeT.failNowed {
		t.Fatal("should have failNowed")
	}
}

func TestErrorFailure(t *testing.T) {
	t.Run("nil error", func(t *testing.T) {
		fakeT := &fakeTestingT{}

		var err error
		assert.Error(fakeT, err, "this error")
		expectFailNowed(t, fakeT, "assertion failed: expected an error, got nil")
	})
	t.Run("different error", func(t *testing.T) {
		fakeT := &fakeTestingT{}

		err := fmt.Errorf("the actual error")
		assert.Error(fakeT, err, "this error")
		expected := `assertion failed: expected error "this error", got "the actual error"`
		expectFailNowed(t, fakeT, expected)
	})
}

func TestErrorContainsFailure(t *testing.T) {
	t.Run("nil error", func(t *testing.T) {
		fakeT := &fakeTestingT{}

		var err error
		assert.ErrorContains(fakeT, err, "this error")
		expectFailNowed(t, fakeT, "assertion failed: expected an error, got nil")
	})
	t.Run("different error", func(t *testing.T) {
		fakeT := &fakeTestingT{}

		err := fmt.Errorf("the actual error")
		assert.ErrorContains(fakeT, err, "this error")
		expected := `assertion failed: expected error to contain "this error", got "the actual error"`
		expectFailNowed(t, fakeT, expected)
	})
}

func TestErrorTypeFailure(t *testing.T) {
	t.Run("nil error", func(t *testing.T) {
		fakeT := &fakeTestingT{}

		var err error
		assert.ErrorType(fakeT, err, os.IsNotExist)
		expectFailNowed(t, fakeT, "assertion failed: error is nil, not os.IsNotExist")
	})
	t.Run("different error", func(t *testing.T) {
		fakeT := &fakeTestingT{}

		err := fmt.Errorf("the actual error")
		assert.ErrorType(fakeT, err, os.IsNotExist)
		expected := `assertion failed: error is the actual error (*errors.errorString), not os.IsNotExist`
		expectFailNowed(t, fakeT, expected)
	})
}
