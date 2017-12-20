package assert

import (
	"fmt"
	"testing"
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

func TestTesterAssertWithBoolFailure(t *testing.T) {
	fakeT := &fakeTestingT{}
	assert := New(fakeT)

	assert.Assert(1 > 5)
	expectFailNowed(t, fakeT, "assertion failed: 1 > 5 is false")

}

func TestTesterAssertWithBoolFailureAndExtraMessage(t *testing.T) {
	fakeT := &fakeTestingT{}
	assert := New(fakeT)

	assert.Assert(1 > 5, "sometimes things fail")
	expectFailNowed(t, fakeT, "assertion failed: 1 > 5 is false: sometimes things fail")
}

func TestTesterAssertWithBoolSuccess(t *testing.T) {
	fakeT := &fakeTestingT{}
	assert := New(fakeT)

	assert.Assert(1 < 5)
	expectSuccess(t, fakeT)
}

func TestTesterAssertWithBoolMultiLineFailure(t *testing.T) {
	fakeT := &fakeTestingT{}
	assert := New(fakeT)

	assert.Assert(func() bool {
		for range []int{1, 2, 3, 4} {
		}
		return false
	}())
	expectFailNowed(t, fakeT, `assertion failed: func() bool {
	for range []int{1, 2, 3, 4} {
	}
	return false
}() is false`)
}

type exampleComparison struct {
	success bool
	message string
}

func (c exampleComparison) Compare() (bool, string) {
	return c.success, c.message
}

func TestTesterAssertWithComparisonSuccess(t *testing.T) {
	fakeT := &fakeTestingT{}
	assert := New(fakeT)

	cmp := exampleComparison{success: true}
	assert.Assert(cmp)
	expectSuccess(t, fakeT)
}

func TestTesterAssertWithComparisonFailure(t *testing.T) {
	fakeT := &fakeTestingT{}
	assert := New(fakeT)

	cmp := exampleComparison{message: "oops, not good"}
	assert.Assert(cmp)
	expectFailNowed(t, fakeT, "assertion failed: oops, not good")
}

func TestTesterAssertWithComparisonAndExtraMessage(t *testing.T) {
	fakeT := &fakeTestingT{}
	assert := New(fakeT)

	cmp := exampleComparison{message: "oops, not good"}
	assert.Assert(cmp, "extra stuff %v", true)
	expectFailNowed(t, fakeT, "assertion failed: oops, not good: extra stuff true")
}

func TestAssertWithBoolFailure(t *testing.T) {
	fakeT := &fakeTestingT{}

	Assert(fakeT, 1 == 6)
	expectFailNowed(t, fakeT, "assertion failed: 1 == 6 is false")
}

type customError struct{}

func (e *customError) Error() string {
	return "custom error"
}

func TestTesterNoErrorSuccess(t *testing.T) {
	fakeT := &fakeTestingT{}
	assert := New(fakeT)

	var err error
	assert.NoError(err)
	expectSuccess(t, fakeT)

	assert.NoError(nil)
	expectSuccess(t, fakeT)

	var customErr *customError
	assert.NoError(customErr)
}

func TestTesterNoErrorBadArg(t *testing.T) {
	fakeT := &fakeTestingT{}
	assert := New(fakeT)

	assert.NoError(3, 4, 5)
	expectFailNowed(t, fakeT, "assertion failed: last argument to NoError() must be an error, got int")
}

func TestTesterNoErrorFailure(t *testing.T) {
	fakeT := &fakeTestingT{}
	assert := New(fakeT)

	assert.NoError(fmt.Errorf("this is the error"))
	expectFailNowed(t, fakeT, "assertion failed: expected no error, got this is the error")
}

func TestTesterNoErrorWithMultiArgFailure(t *testing.T) {
	fakeT := &fakeTestingT{}
	assert := New(fakeT)

	assert.NoError(func() (bool, int, error) {
		return true, 3, fmt.Errorf("this is the error")
	}())
	expectFailNowed(t, fakeT, "assertion failed: expected no error, got this is the error")
}

func TestTesterCheckFailure(t *testing.T) {
	fakeT := &fakeTestingT{}
	assert := New(fakeT)

	if assert.Check(1 == 2) {
		t.Error("expected check to return false on failure")
	}
	expectFailed(t, fakeT, "assertion failed: 1 == 2 is false")
}

func TestTesterCheckSuccess(t *testing.T) {
	fakeT := &fakeTestingT{}
	assert := New(fakeT)

	if !assert.Check(1 == 1) {
		t.Error("expected check to return true on success")
	}
	expectSuccess(t, fakeT)
}

func TestTesterEqualSuccess(t *testing.T) {
	fakeT := &fakeTestingT{}
	assert := New(fakeT)

	assert.Equal(1, 1)
	expectSuccess(t, fakeT)

	assert.Equal("abcd", "abcd")
	expectSuccess(t, fakeT)
}

func TestTesterEqualFailure(t *testing.T) {
	fakeT := &fakeTestingT{}
	assert := New(fakeT)

	assert.Equal(1, 3)
	expectFailNowed(t, fakeT, "assertion failed: 1 (int) != 3 (int)")
}

func TestTesterEqualFailureTypes(t *testing.T) {
	fakeT := &fakeTestingT{}
	assert := New(fakeT)

	assert.Equal(3, "3")
	expectFailNowed(t, fakeT, `assertion failed: 3 (int) != 3 (string)`)
}

type testingT interface {
	Errorf(msg string, args ...interface{})
	Fatalf(msg string, args ...interface{})
}

func expectFailNowed(t testingT, fakeT *fakeTestingT, expected string) {
	if ht, ok := t.(helperT); ok {
		ht.Helper()
	}
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
	if ht, ok := t.(helperT); ok {
		ht.Helper()
	}
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
	if ht, ok := t.(helperT); ok {
		ht.Helper()
	}
	if fakeT.failNowed {
		t.Errorf("should not have failNowed, got messages %s", fakeT.msgs)
	}
	if fakeT.failed {
		t.Errorf("should not have failed, got messages %s", fakeT.msgs)
	}
}
