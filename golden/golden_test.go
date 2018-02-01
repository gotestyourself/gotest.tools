package golden

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/gotestyourself/gotestyourself/assert"
	"github.com/gotestyourself/gotestyourself/assert/cmp"
	"github.com/gotestyourself/gotestyourself/fs"
)

type fakeT struct {
	Failed bool
}

func (t *fakeT) Log(...interface{}) {
}

func (t *fakeT) FailNow() {
	t.Failed = true
}

func (t *fakeT) Fail() {
	t.Failed = true
}

func TestGoldenGetInvalidFile(t *testing.T) {
	fakeT := new(fakeT)

	Get(fakeT, "/invalid/path")
	assert.Assert(t, fakeT.Failed)
}

func TestGoldenGetAbsolutePath(t *testing.T) {
	file := fs.NewFile(t, "abs-test", fs.WithContent("content\n"))
	defer file.Remove()
	fakeT := new(fakeT)

	Get(fakeT, file.Path())
	assert.Assert(t, !fakeT.Failed)
}

func TestGoldenGet(t *testing.T) {
	expected := "content\nline1\nline2"

	filename, clean := setupGoldenFile(t, expected)
	defer clean()

	fakeT := new(fakeT)

	actual := Get(fakeT, filename)
	assert.Assert(t, !fakeT.Failed)
	assert.Assert(t, cmp.DeepEqual(actual, []byte(expected)))
}

func TestGoldenAssertInvalidContent(t *testing.T) {
	filename, clean := setupGoldenFile(t, "content")
	defer clean()

	fakeT := new(fakeT)

	success := Assert(fakeT, "foo", filename)
	assert.Assert(t, fakeT.Failed)
	assert.Assert(t, !success)
}

func TestGoldenAssertInvalidContentUpdate(t *testing.T) {
	undo := setUpdateFlag()
	defer undo()
	filename, clean := setupGoldenFile(t, "content")
	defer clean()

	fakeT := new(fakeT)

	success := Assert(fakeT, "foo", filename)
	assert.Assert(t, !fakeT.Failed)
	assert.Assert(t, success)
}

func TestGoldenAssert(t *testing.T) {
	filename, clean := setupGoldenFile(t, "foo")
	defer clean()

	fakeT := new(fakeT)

	success := Assert(fakeT, "foo", filename)
	assert.Assert(t, !fakeT.Failed)
	assert.Assert(t, success)
}

func TestGoldenAssertBytes(t *testing.T) {
	filename, clean := setupGoldenFile(t, "foo")
	defer clean()

	fakeT := new(fakeT)

	success := AssertBytes(fakeT, []byte("foo"), filename)
	assert.Assert(t, !fakeT.Failed)
	assert.Assert(t, success)
}

func setUpdateFlag() func() {
	oldFlagUpdate := *flagUpdate
	*flagUpdate = true
	return func() { *flagUpdate = oldFlagUpdate }
}

func setupGoldenFile(t *testing.T, content string) (string, func()) {
	_ = os.Mkdir("testdata", 0755)
	f, err := ioutil.TempFile("testdata", "")
	assert.Assert(t, err, "fail to setup test golden file")
	err = ioutil.WriteFile(f.Name(), []byte(content), 0660)
	assert.Assert(t, err, "fail to write test golden file with %q", content)
	_, name := filepath.Split(f.Name())
	t.Log(f.Name(), name)
	return name, func() {
		assert.NilError(t, os.Remove(f.Name()))
	}
}

func TestStringFailure(t *testing.T) {
	filename, clean := setupGoldenFile(t, "this is\nthe text")
	defer clean()

	result := String("this is\nnot the text", filename)()
	assert.Assert(t, !result.Success())
	assert.Equal(t, result.(failure).FailureMessage(), `
--- expected
+++ actual
@@ -1,2 +1,2 @@
 this is
-the text
+not the text
`)
}

type failure interface {
	FailureMessage() string
}

func TestBytesFailure(t *testing.T) {
	filename, clean := setupGoldenFile(t, "5556")
	defer clean()

	result := Bytes([]byte("5555"), filename)()
	assert.Assert(t, !result.Success())
	assert.Equal(t, result.(failure).FailureMessage(),
		`[53 53 53 53] (actual) != [53 53 53 54] (expected)`)
}
