package poll

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"gotest.tools/v3/assert"
)

func TestWaitOnFile(t *testing.T) {
	fakeFilePath := "./fakefile"

	check := FileExists(fakeFilePath)

	ctx := context.Background()
	t.Run("file does not exist", func(t *testing.T) {
		err := check(ctx, t)
		assert.Assert(t, errors.As(err, &cont{}))
		assert.Error(t, err, fmt.Sprintf("file %s does not exist", fakeFilePath))
	})

	os.Create(fakeFilePath)
	defer os.Remove(fakeFilePath)

	t.Run("file exists", func(t *testing.T) {
		assert.NilError(t, check(ctx, t))
	})
}

func TestWaitOnSocketWithTimeout(t *testing.T) {
	ctx := context.Background()
	t.Run("connection to unavailable address", func(t *testing.T) {
		check := Connection("tcp", "foo.bar:55555")
		err := check(ctx, t)
		assert.Assert(t, errors.As(err, &cont{}))
		assert.Error(t, err, "socket tcp://foo.bar:55555 not available")
	})

	t.Run("connection to ", func(t *testing.T) {
		check := Connection("tcp", "google.com:80")
		assert.Assert(t, !errors.As(check(ctx, t), &cont{}))
	})
}
