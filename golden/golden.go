// Package golden provides function and helpers to use golden file for
// testing purpose.
package golden

import (
	"flag"
	"io/ioutil"
	"path/filepath"

	"github.com/stretchr/testify/assert"
)

type testingT interface {
	Fatalf(string, ...interface{})
	Fatal(...interface{})
	Errorf(string, ...interface{})
}

var update = flag.Bool("test.update", false, "update golden file")

// Get returns the golden file content. If the `test.update` is specified, it updates the
// file with the current output and returns it.
func Get(t testingT, actual []byte, filename string) []byte {
	golden := filepath.Join("testdata", filename)
	if *update {
		if err := ioutil.WriteFile(golden, actual, 0644); err != nil {
			t.Fatal(err)
		}
	}
	expected, err := ioutil.ReadFile(golden)
	if err != nil {
		t.Fatal(err)
	}
	return expected
}

// Assert asserts that the actual content and the golden file are equal.
//
//    golden.Assert(t, []byte("foo"), "testdata/foo-content")
//
// Returns whether the assertion was successful (true) or not (false)
func Assert(t testingT, actual []byte, filename string) bool {
	expected := Get(t, actual, filename)
	return assert.Equal(t, expected, actual)
}
