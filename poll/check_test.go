package poll

import (
	"fmt"
	"net"
	"os"
	"testing"

	"gotest.tools/v3/assert"
)

func TestWaitOnFile(t *testing.T) {
	fakeFilePath := "./fakefile"

	check := FileExists(fakeFilePath)

	t.Run("file does not exist", func(t *testing.T) {
		r := check(t)
		assert.Assert(t, !r.Done())
		assert.Equal(t, r.Message(), fmt.Sprintf("file %s does not exist", fakeFilePath))
	})

	os.Create(fakeFilePath)
	defer os.Remove(fakeFilePath)

	t.Run("file exists", func(t *testing.T) {
		assert.Assert(t, check(t).Done())
	})
}

func TestWaitOnSocketWithTimeout(t *testing.T) {
	t.Run("connection to unavailable address", func(t *testing.T) {
		check := Connection("tcp", "foo.bar:55555")
		r := check(t)
		assert.Assert(t, !r.Done())
		assert.Equal(t, r.Message(), "socket tcp://foo.bar:55555 not available")
	})

	t.Run("connection to available address", func(t *testing.T) {
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		assert.NilError(t, err)
		defer listener.Close()

		check := Connection("tcp", listener.Addr().String())
		assert.Assert(t, check(t).Done())
	})
}
