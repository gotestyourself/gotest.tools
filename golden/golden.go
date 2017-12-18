/*Package golden provides tools for comparing large mutli-line strings.

Golden files are files in the ./testdata/ subdirectory of the package under test.
*/
package golden

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/gotestyourself/gotestyourself/assert"
	"github.com/gotestyourself/gotestyourself/assert/cmp"
	"github.com/gotestyourself/gotestyourself/internal/format"
	"github.com/pmezard/go-difflib/difflib"
)

var flagUpdate = flag.Bool("test.update-golden", false, "update golden file")

type helperT interface {
	Helper()
}

// Get returns the golden file content
func Get(t assert.TestingT, filename string) []byte {
	if ht, ok := t.(helperT); ok {
		ht.Helper()
	}
	expected, err := ioutil.ReadFile(Path(filename))
	assert.NoError(t, err)
	return expected
}

// Path returns the full path to a golden file
func Path(filename string) string {
	if filepath.IsAbs(filename) {
		return filename
	}
	return filepath.Join("testdata", filename)
}

func update(t assert.TestingT, filename string, actual []byte) {
	if *flagUpdate {
		err := ioutil.WriteFile(Path(filename), actual, 0644)
		assert.NoError(t, err)
	}
}

// Assert compares the actual content to the expected content in the golden file.
// If the `-test.update-golden` flag is set then the actual content is written
// to the golden file.
// Returns whether the assertion was successful (true) or not (false)
func Assert(t assert.TestingT, actual string, filename string, msgAndArgs ...interface{}) bool {
	if ht, ok := t.(helperT); ok {
		ht.Helper()
	}
	update(t, filename, []byte(actual))
	expected := Get(t, filename)

	if bytes.Equal(expected, []byte(actual)) {
		return true
	}

	diff, err := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
		A:        difflib.SplitLines(string(expected)),
		B:        difflib.SplitLines(actual),
		FromFile: "Expected",
		ToFile:   "Actual",
		Context:  3,
	})
	assert.Assert(t, cmp.NoError(err), msgAndArgs...)
	t.Log(format.WithCustomMessage(fmt.Sprintf("Not Equal: \n%s", diff), msgAndArgs...))
	t.Fail()
	return false
}

// AssertBytes compares the actual result to the expected result in the golden
// file. If the `-test.update-golden` flag is set then the actual content is
// written to the golden file.
// Returns whether the assertion was successful (true) or not (false)
// nolint: lll
func AssertBytes(t assert.TestingT, actual []byte, filename string, msgAndArgs ...interface{}) bool {
	if ht, ok := t.(helperT); ok {
		ht.Helper()
	}
	update(t, filename, actual)
	expected := Get(t, filename)
	return assert.Check(t, cmp.Compare(expected, actual), msgAndArgs...)
}
