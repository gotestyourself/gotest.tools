package assert_test

import (
	"fmt"
	"testing"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/internal/source"
)

func TestEqual_WithGoldenUpdate(t *testing.T) {
	t.Run("assert failed with update=false", func(t *testing.T) {
		ft := &fakeTestingT{}
		actual := `not this value`
		assert.Equal(ft, actual, expectedOne)
		assert.Assert(t, ft.failNowed)
	})

	t.Run("value is updated when -update=true", func(t *testing.T) {
		patchUpdate(t)
		ft := &fakeTestingT{}

		actual := `this is the
actual value
that we are testing against`
		assert.Equal(ft, actual, expectedOne)

		// reset
		fmt.Println("WHHHHHHHHHHY")
		assert.Equal(ft, "\n\n\n", expectedOne)
	})
}

var expectedOne = `


`

func patchUpdate(t *testing.T) {
	source.Update = true
	t.Cleanup(func() {
		source.Update = false
	})
}

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
