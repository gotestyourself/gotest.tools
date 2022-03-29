//go:build go1.18
// +build go1.18

package assert

// TestingT is the subset of testing.T used by the assert package.
type TestingT interface {
	FailNow()
	Fail()
	Log(args ...any)
}
