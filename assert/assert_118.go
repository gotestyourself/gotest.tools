//go:build go1.18
// +build go1.18

package assert

import (
	gocmp "github.com/google/go-cmp/cmp"
	"gotest.tools/v3/assert/cmp"
	"gotest.tools/v3/internal/assert"
)

// Equal uses the == operator to assert two values are equal and fails the test
// if they are not equal.
//
// If the comparison fails Equal will use the variable names and types of
// x and y as part of the failure message to identify the actual and expected
// values.
//
//   assert.Equal(t, actual, expected)
//   // main_test.go:41: assertion failed: 1 (actual int) != 21 (expected int32)
//
// If either x or y are a multi-line string the failure message will include a
// unified diff of the two values. If the values only differ by whitespace
// the unified diff will be augmented by replacing whitespace characters with
// visible characters to identify the whitespace difference.
//
// Equal uses t.FailNow to fail the test. Like t.FailNow, Equal must be
// called from the goroutine running the test function, not from other
// goroutines created during the test. Use Check with cmp.Equal from other
// goroutines.
func Equal[T comparable](t TestingT, x, y T, msgAndArgs ...interface{}) {
	if ht, ok := t.(helperT); ok {
		ht.Helper()
	}
	if !assert.Eval(t, assert.ArgsAfterT, cmp.Equal(x, y), msgAndArgs...) {
		t.FailNow()
	}
}

// DeepEqual uses google/go-cmp (https://godoc.org/github.com/google/go-cmp/cmp)
// to assert two values are equal and fails the test if they are not equal.
//
// Package http://pkg.go.dev/gotest.tools/v3/assert/opt provides some additional
// commonly used Options.
//
// DeepEqual uses t.FailNow to fail the test. Like t.FailNow, DeepEqual must be
// called from the goroutine running the test function, not from other
// goroutines created during the test. Use Check with cmp.DeepEqual from other
// goroutines.
func DeepEqual[T any](t TestingT, x, y T, opts ...gocmp.Option) {
	if ht, ok := t.(helperT); ok {
		ht.Helper()
	}
	if !assert.Eval(t, assert.ArgsAfterT, cmp.DeepEqual(x, y, opts...)) {
		t.FailNow()
	}
}
