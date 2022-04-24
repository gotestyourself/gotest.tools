//go:build !go1.18
// +build !go1.18

package assert

import (
	gocmp "github.com/google/go-cmp/cmp"
	"gotest.tools/v3/assert/cmp"
	"gotest.tools/v3/internal/assert"
)

func Equal(t TestingT, x, y interface{}, msgAndArgs ...interface{}) {
	if ht, ok := t.(helperT); ok {
		ht.Helper()
	}
	if !assert.Eval(t, assert.ArgsAfterT, cmp.Equal(x, y), msgAndArgs...) {
		t.FailNow()
	}
}

func DeepEqual(t TestingT, x, y interface{}, opts ...gocmp.Option) {
	if ht, ok := t.(helperT); ok {
		ht.Helper()
	}
	if !assert.Eval(t, assert.ArgsAfterT, cmp.DeepEqual(x, y, opts...)) {
		t.FailNow()
	}
}
