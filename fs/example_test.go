package fs

import (
	"io/ioutil"

	"os"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeTesting struct{}

func (t fakeTesting) Errorf(format string, args ...interface{}) {}
func (t fakeTesting) FailNow()                                  {}

var t = fakeTesting{}

// Create a temporary directory which contains a single file
func ExampleNewDir() {
	dir := NewDir(t, "test-name", WithFile("file1", "content\n"))
	defer dir.Remove()

	files, err := ioutil.ReadDir(dir.Path())
	require.NoError(t, err)
	assert.Len(t, files, 0)
}

// Create a new file with some content
func ExampleNewFile() {
	file := NewFile(t, "test-name", WithContent("content\n"), AsUser(0, 0))
	defer file.Remove()

	content, err := ioutil.ReadFile(file.Path())
	require.NoError(t, err)
	assert.Equal(t, "content\n", content)
}

// Create a directory and subdirectory with files
func ExampleWithDir() {
	dir := NewDir(t, "test-name",
		WithDir("subdir",
			WithMode(os.FileMode(0700)),
			WithFile("file1", "content\n")),
	)
	defer dir.Remove()
}
