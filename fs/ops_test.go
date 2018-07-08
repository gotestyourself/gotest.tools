package fs_test

import (
	"runtime"
	"testing"

	"gotest.tools/assert"
	"gotest.tools/fs"
	"gotest.tools/skip"
)

func TestFromDir(t *testing.T) {
	dir := fs.NewDir(t, "test-from-dir", fs.FromDir("testdata/copy-test"))
	defer dir.Remove()

	expected := fs.Expected(t,
		fs.WithFile("1", "1\n"),
		fs.WithDir("a",
			fs.WithFile("1", "1\n"),
			fs.WithFile("2", "2\n"),
			fs.WithDir("b",
				fs.WithFile("1", "1\n"))))

	assert.Assert(t, fs.Equal(dir.Path(), expected))
}

func TestFromDirSymlink(t *testing.T) {
	skip.If(t, runtime.GOOS == "windows", "See gotest.tools/issues/107")
	dir := fs.NewDir(t, "test-from-dir", fs.FromDir("testdata/copy-test-with-symlink"))
	defer dir.Remove()

	expected := fs.Expected(t,
		fs.WithFile("1", "1\n"),
		fs.WithDir("a",
			fs.WithFile("1", "1\n"),
			fs.WithFile("2", "2\n"),
			fs.WithDir("b",
				fs.WithFile("1", "1\n"),
				fs.WithSymlink("2", "../2"),
				fs.WithSymlink("3", "/some/inexistent/link"),
				fs.WithSymlink("4", "5"),
			)))

	assert.Assert(t, fs.Equal(dir.Path(), expected))
}
