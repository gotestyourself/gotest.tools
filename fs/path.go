package fs

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"

	"github.com/gotestyourself/gotestyourself/assert"
)

// resourcePath is an adaptor for resources so they can be used as a Path
// with PathOps.
type resourcePath struct{}

func (p *resourcePath) Path() string {
	return "manifest: not a filesystem path"
}

func (p *resourcePath) Remove() {}

type filePath struct {
	resourcePath
	file *file
}

func (p *filePath) SetContent(content io.ReadCloser) {
	p.file.content = content
}

func (p *filePath) SetUID(uid uint32) {
	p.file.uid = uid
}

func (p *filePath) SetGID(gid uint32) {
	p.file.gid = gid
}

type directoryPath struct {
	resourcePath
	directory *directory
}

func (p *directoryPath) SetUID(uid uint32) {
	p.directory.uid = uid
}

func (p *directoryPath) SetGID(gid uint32) {
	p.directory.gid = gid
}

func (p *directoryPath) AddSymlink(path, target string) error {
	p.directory.items[path] = &symlink{
		resource: newResource(defaultSymlinkMode),
		target:   target,
	}
	return nil
}

func (p *directoryPath) AddFile(path string, ops ...PathOp) error {
	newFile := &file{resource: newResource(0)}
	p.directory.items[path] = newFile
	exp := &filePath{file: newFile}
	return applyPathOps(exp, ops)
}

func (p *directoryPath) AddDirectory(path string, ops ...PathOp) error {
	newDir := newDirectoryWithDefaults()
	p.directory.items[path] = newDir
	exp := &directoryPath{directory: newDir}
	return applyPathOps(exp, ops)
}

// Expected returns a Manifest with a directory structured created by ops.
func Expected(t assert.TestingT, ops ...PathOp) Manifest {
	if ht, ok := t.(helperT); ok {
		ht.Helper()
	}

	newDir := newDirectoryWithDefaults()
	e := &directoryPath{directory: newDir}
	assert.NilError(t, applyPathOps(e, ops))
	return Manifest{root: newDir}
}

func newDirectoryWithDefaults() *directory {
	return &directory{
		resource: newResource(defaultRootDirMode),
		items:    make(map[string]dirEntry),
	}
}

func newResource(mode os.FileMode) resource {
	return resource{
		mode: mode,
		uid:  currentUID(),
		gid:  currentGID(),
	}
}

func currentUID() uint32 {
	return normalizeID(os.Getuid())
}

func currentGID() uint32 {
	return normalizeID(os.Getgid())
}

func normalizeID(id int) uint32 {
	// ids will be -1 on windows
	if id < 0 {
		return 0
	}
	return uint32(id)
}

var anyFileContent = ioutil.NopCloser(bytes.NewReader(nil))

// AllowAnyFileContent is a PathOp that sets the file entry in the Manifest to
// match against any file content.
func AllowAnyFileContent(path Path) error {
	if m, ok := path.(*filePath); ok {
		m.SetContent(anyFileContent)
	}
	return nil
}

const anyFile = "*"

// AllowExtraFiles is a PathOp that updates the Manifest to allow a directory
// to contain unspecified files.
func AllowExtraFiles(path Path) error {
	if m, ok := path.(*directoryPath); ok {
		m.AddFile(anyFile)
	}
	return nil
}

// TODO: AllowExtraFilesWith(ops ...PathOp) PathOp {}
