package fs_test

import (
	"io/ioutil"
	"os"
	"testing"

	"gotest.tools/assert"
	"gotest.tools/fs"
)

func TestNewDirWithOpsAndManifestEqual(t *testing.T) {
	var userOps []fs.PathOp
	if os.Geteuid() == 0 {
		userOps = append(userOps, fs.AsUser(1001, 1002))
	}

	ops := []fs.PathOp{
		fs.WithFile("file1", "contenta", fs.WithMode(0400)),
		fs.WithFile("file2", "", fs.WithBytes([]byte{0, 1, 2})),
		fs.WithFile("file5", "", userOps...),
		fs.WithSymlink("link1", "file1"),
		fs.WithDir("sub",
			fs.WithFiles(map[string]string{
				"file3": "contentb",
				"file4": "contentc",
			}),
			fs.WithMode(0705),
		),
	}

	dir := fs.NewDir(t, "test-all", ops...)
	defer dir.Remove()

	manifestOps := append(
		ops[:3],
		fs.WithSymlink("link1", dir.Join("file1")),
		ops[4],
	)
	assert.Assert(t, fs.Equal(dir.Path(), fs.Expected(t, manifestOps...)))
}

func TestNewFile(t *testing.T) {
	t.Run("with test name", func(t *testing.T) {
		tmpFile := fs.NewFile(t, t.Name())
		_, err := os.Stat(tmpFile.Path())
		assert.NilError(t, err)

		tmpFile.Remove()
		_, err = os.Stat(tmpFile.Path())
		assert.ErrorType(t, err, os.IsNotExist)
	})

	t.Run(`with \ in name`, func(t *testing.T) {
		tmpFile := fs.NewFile(t, `foo\thing`)
		_, err := os.Stat(tmpFile.Path())
		assert.NilError(t, err)

		tmpFile.Remove()
		_, err = os.Stat(tmpFile.Path())
		assert.ErrorType(t, err, os.IsNotExist)
	})

}

func TestUpdate(t *testing.T) {
	t.Run("with file", func(t *testing.T) {
		tmpFile := fs.NewFile(t, "test-update-file", fs.WithContent("contenta"))
		defer tmpFile.Remove()
		tmpFile.Update(t, fs.WithContent("contentb"))
		content, err := ioutil.ReadFile(tmpFile.Path())
		assert.NilError(t, err)
		assert.Equal(t, string(content), "contentb")
	})

	t.Run("with dir", func(t *testing.T) {
		tmpDir := fs.NewDir(t, "test-update-dir")
		defer tmpDir.Remove()
		tmpDir.Update(t, fs.WithFile("file1", "contenta"))
		tmpDir.Update(t, fs.WithFile("file2", "contentb"))
		expected := fs.Expected(t,
			fs.WithFile("file1", "contenta"),
			fs.WithFile("file2", "contentb"))
		assert.Assert(t, fs.Equal(tmpDir.Path(), expected))
	})
}
