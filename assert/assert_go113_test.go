// +build go1.13

package assert

import (
	"fmt"
	"os"
	"testing"
)

func TestErrorIs(t *testing.T) {
	t.Run("nil error", func(t *testing.T) {
		fakeT := &fakeTestingT{}

		var err error
		ErrorIs(fakeT, err, os.ErrNotExist)
		expected := `assertion failed: error is nil, not "file does not exist" (os.ErrNotExist *errors.errorString)`
		expectFailNowed(t, fakeT, expected)
	})
	t.Run("different error", func(t *testing.T) {
		fakeT := &fakeTestingT{}

		err := fmt.Errorf("the actual error")
		ErrorIs(fakeT, err, os.ErrNotExist)
		expected := `assertion failed: error is "the actual error" (err *errors.errorString), not "file does not exist" (os.ErrNotExist *errors.errorString)`
		expectFailNowed(t, fakeT, expected)
	})
	t.Run("same error", func(t *testing.T) {
		fakeT := &fakeTestingT{}

		err := fmt.Errorf("some wrapping: %w", os.ErrNotExist)
		ErrorIs(fakeT, err, os.ErrNotExist)
		expectSuccess(t, fakeT)
	})
}
