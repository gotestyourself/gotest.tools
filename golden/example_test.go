package golden_test

import (
	"testing"

	"github.com/gotestyourself/gotestyourself/golden"
)

var t = &testing.T{}

func ExampleAssert() {
	golden.Assert(t, "foo", "foo-content.golden")
}

func ExampleAssertBytes() {
	golden.AssertBytes(t, []byte("foo"), "foo-content.golden")
}
