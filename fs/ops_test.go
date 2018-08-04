package fs_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"gotest.tools/assert"
	"gotest.tools/fs"
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
	dir := fs.NewDir(t, "test-from-dir", fs.FromDir("testdata/copy-test-with-symlink"))
	defer dir.Remove()

	currentdir, err := os.Getwd()
	assert.NilError(t, err)

	link2 := filepath.FromSlash("../2")
	link3 := "/some/inexistent/link"
	if runtime.GOOS == "windows" {
		link3 = filepath.Join(filepath.VolumeName(currentdir), link3)
	}

	expected := fs.Expected(t,
		fs.WithFile("1", "1\n"),
		fs.WithDir("a",
			fs.WithFile("1", "1\n"),
			fs.WithFile("2", "2\n"),
			fs.WithDir("b",
				fs.WithFile("1", "1\n"),
				fs.WithSymlink("2", link2),
				fs.WithSymlink("3", link3),
				fs.WithSymlink("4", "5"),
			)))

	assert.Assert(t, fs.Equal(dir.Path(), expected))
}

func TestWithTimestamps(t *testing.T) {
	stamp := time.Date(2011, 11, 11, 5, 55, 55, 0, time.UTC)
	tmpFile := fs.NewFile(t, t.Name(), fs.WithTimestamps(stamp, stamp))
	defer tmpFile.Remove()

	stat, err := os.Stat(tmpFile.Path())
	assert.NilError(t, err)
	assert.DeepEqual(t, stat.ModTime(), stamp)
}
