package fs_test

import (
	"io/ioutil"
	"os"

	"github.com/gotestyourself/gotestyourself/fs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeTesting struct{}

func (t fakeTesting) Errorf(format string, args ...interface{}) {}
func (t fakeTesting) FailNow()                                  {}

var t = fakeTesting{}

// Create a temporary directory which contains a single file
func ExampleNewDir() {
	dir := fs.NewDir(t, "test-name", fs.WithFile("file1", "content\n"))
	defer dir.Remove()

	files, err := ioutil.ReadDir(dir.Path())
	require.NoError(t, err)
	assert.Len(t, files, 0)
}

// Create a new file with some content
func ExampleNewFile() {
	file := fs.NewFile(t, "test-name", fs.WithContent("content\n"), fs.AsUser(0, 0))
	defer file.Remove()

	content, err := ioutil.ReadFile(file.Path())
	require.NoError(t, err)
	assert.Equal(t, "content\n", content)
}

// Create a directory and subdirectory with files
func ExampleWithDir() {
	dir := fs.NewDir(t, "test-name",
		fs.WithDir("subdir",
			fs.WithMode(os.FileMode(0700)),
			fs.WithFile("file1", "content\n")),
	)
	defer dir.Remove()
}
