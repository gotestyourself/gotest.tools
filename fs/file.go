/*Package fs provides tools for creating temporary files, and testing the
contents and structure of a directory.
*/
package fs // import "gotest.tools/fs"

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"gotest.tools/assert"
	"gotest.tools/x/subtest"
)

// Path objects return their filesystem path. Path may be implemented by a
// real filesystem object (such as File and Dir) or by a type which updates
// entries in a Manifest.
type Path interface {
	Path() string
	Remove()
}

var (
	_ Path = &Dir{}
	_ Path = &File{}
)

// File is a temporary file on the filesystem
type File struct {
	path string
}

type helperT interface {
	Helper()
}

// NewFile creates a new file in a temporary directory using prefix as part of
// the filename. The PathOps are applied to the before returning the File.
func NewFile(t assert.TestingT, prefix string, ops ...PathOp) *File {
	if ht, ok := t.(helperT); ok {
		ht.Helper()
	}
	tempfile, err := ioutil.TempFile("", cleanPrefix(prefix)+"-")
	assert.NilError(t, err)
	return wrapFile(t, tempfile, ops...)
}

func cleanPrefix(prefix string) string {
	// windows requires both / and \ are replaced
	if runtime.GOOS == "windows" {
		prefix = strings.Replace(prefix, string(os.PathSeparator), "-", -1)
	}
	return strings.Replace(prefix, "/", "-", -1)
}

// Path returns the full path to the file
func (f *File) Path() string {
	return f.path
}

// Remove the file
func (f *File) Remove() {
	// nolint: errcheck
	os.Remove(f.path)
}

// Dir is a temporary directory
type Dir struct {
	path string
}

// NewDir returns a new temporary directory using prefix as part of the directory
// name. The PathOps are applied before returning the Dir.
func NewDir(t assert.TestingT, prefix string, ops ...PathOp) *Dir {
	if ht, ok := t.(helperT); ok {
		ht.Helper()
	}
	path, err := ioutil.TempDir("", cleanPrefix(prefix)+"-")
	assert.NilError(t, err)
	return wrapDir(t, path, ops...)
}

// NewFile creates a new file in the directory with the specified name.
// The PathOps are applied to the before returning the File.
func (d *Dir) NewFile(t assert.TestingT, name string, ops ...PathOp) *File {
	if ht, ok := t.(helperT); ok {
		ht.Helper()
	}
	f, err := os.Create(d.Join(name))
	assert.NilError(t, err)
	return wrapFile(t, f, ops...)
}

// NewDir returns a new subdirectory in the directory with the specified name.
// The PathOps are applied before returning the Dir.
func (d *Dir) NewDir(t assert.TestingT, name string, ops ...PathOp) *Dir {
	if ht, ok := t.(helperT); ok {
		ht.Helper()
	}
	path := d.Join(name)
	err := os.Mkdir(path, os.ModePerm)
	assert.NilError(t, err)
	return wrapDir(t, path, ops...)
}

func wrapDir(t assert.TestingT, path string, ops ...PathOp) *Dir {
	dir := &Dir{path: path}
	for _, op := range ops {
		assert.NilError(t, op(dir))
	}
	if tc, ok := t.(subtest.TestContext); ok {
		tc.AddCleanup(dir.Remove)
	}
	return dir
}

func wrapFile(t assert.TestingT, f *os.File, ops ...PathOp) *File {
	file := &File{path: f.Name()}
	assert.NilError(t, f.Close())
	for _, op := range ops {
		assert.NilError(t, op(file))
	}
	if tc, ok := t.(subtest.TestContext); ok {
		tc.AddCleanup(file.Remove)
	}
	return file
}

// Path returns the full path to the directory
func (d *Dir) Path() string {
	return d.path
}

// Remove the directory
func (d *Dir) Remove() {
	// nolint: errcheck
	os.RemoveAll(d.path)
}

// Join returns a new path with this directory as the base of the path
func (d *Dir) Join(parts ...string) string {
	return filepath.Join(append([]string{d.Path()}, parts...)...)
}
