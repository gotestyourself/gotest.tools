package fs_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/fs"
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

	link2 := filepath.FromSlash("../2")
	link3 := filepath.FromSlash("/some/inexistent/link")

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

func TestExplicitSymlinkTargets(t *testing.T) {
	dir := fs.NewDir(t, t.Name())
	defer dir.Remove()

	absolute := dir.Join("target")
	relative := filepath.Join("..", "target")
	contents := []fs.PathOp{
		fs.WithFile("target", "data"),
		fs.WithDir("links",
			fs.WithRelativeSymlink("relative", "../target"),
			fs.WithAbsoluteSymlink("absolute", absolute),
			fs.WithRelativeSymlink("dangling", "missing"),
		),
	}

	fs.Apply(t, dir, contents...)
	assert.Assert(t, fs.Equal(dir.Path(), fs.Expected(t, contents...)))
	for link, want := range map[string]string{"relative": relative, "absolute": absolute, "dangling": "missing"} {
		got, err := os.Readlink(dir.Join("links", link))
		assert.NilError(t, err)
		assert.Equal(t, got, want)
	}
}

func TestExplicitSymlinkTargetKind(t *testing.T) {
	dir := fs.NewDir(t, t.Name())
	defer dir.Remove()
	assert.ErrorContains(t, fs.WithAbsoluteSymlink("link", "relative")(dir), "must be absolute")
	assert.ErrorContains(t, fs.WithRelativeSymlink("link", dir.Join("target"))(dir), "must be relative")
}

func TestWithTimestamps(t *testing.T) {
	stamp := time.Date(2011, 11, 11, 5, 55, 55, 0, time.UTC)
	tmpFile := fs.NewFile(t, t.Name(), fs.WithTimestamps(stamp, stamp))
	defer tmpFile.Remove()

	stat, err := os.Stat(tmpFile.Path())
	assert.NilError(t, err)
	assert.DeepEqual(t, stat.ModTime(), stamp)
}

func TestApply(t *testing.T) {
	t.Run("with file", func(t *testing.T) {
		tmpFile := fs.NewFile(t, "test-update-file", fs.WithContent("contenta"))
		defer tmpFile.Remove()
		fs.Apply(t, tmpFile, fs.WithContent("contentb"))
		content, err := os.ReadFile(tmpFile.Path())
		assert.NilError(t, err)
		assert.Equal(t, string(content), "contentb")
	})

	t.Run("with dir", func(t *testing.T) {
		tmpDir := fs.NewDir(t, "test-update-dir")
		defer tmpDir.Remove()
		fs.Apply(t, tmpDir, fs.WithFile("file1", "contenta"))
		fs.Apply(t, tmpDir, fs.WithFile("file2", "contentb"))
		expected := fs.Expected(t,
			fs.WithFile("file1", "contenta"),
			fs.WithFile("file2", "contentb"))
		assert.Assert(t, fs.Equal(tmpDir.Path(), expected))
	})
}

func TestWithReaderContent(t *testing.T) {
	content := "this is a test"
	dir := fs.NewDir(t, t.Name(),
		fs.WithFile("1", "",
			fs.WithReaderContent(strings.NewReader(content))),
	)
	defer dir.Remove()
	expected := fs.Expected(t, fs.WithFile("1", content))
	assert.Assert(t, fs.Equal(dir.Path(), expected))
}
