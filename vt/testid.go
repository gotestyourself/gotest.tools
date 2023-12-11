package vt

import (
	"fmt"
	"go/token"
	"path/filepath"
	"runtime"
)

// ID creates a [TestID] from the name and position of the caller in the source
// file. ID is used to improve the navigation of large table tests by adding
// a file:line message to the output of a test. This output becomes a hyperlink
// in most IDE.
//
//	type testCase struct {
//		id vt.TestID
//		...
//	}
//
//	for _, tc := range []testCase{...} {
//		t.Run(tc.id.Name, func(t *testing.T) {
//			tc.id.PrintPosition()
//			...
//		})
//	}
func ID(name string) TestID {
	_, filename, line, ok := runtime.Caller(1)
	if !ok {
		panic("failed to get call stack")
	}
	return TestID{
		Name: name,
		position: token.Position{
			Filename: filepath.Base(filename),
			Line:     line,
		},
	}
}

// TestID identifies a test case in a table test by name and file:line position
// in the source file. See [ID] for usage.
type TestID struct {
	Name     string
	position token.Position
}

func (i TestID) PrintPosition() {
	fmt.Printf("    %v: test case: %v\n", i.position, i.Name)
}
