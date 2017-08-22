package golden_test

import (
	"testing"

	"github.com/gotestyourself/gotestyourself/golden"
)

func ExampleAssert() {
	golden.Assert(&testing.T{}, "foo", "foo-content.golden")
}

func ExampleAssertBytes() {
	golden.AssertBytes(&testing.T{}, []byte("foo"), "foo-content.golden")
}
