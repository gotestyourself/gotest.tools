/*Package suite provides compatibility with testify/suite.

Suites can be used to group tests together, and to perform common setup and
teardown for each test in the suite.

TODO: example

Regular expression to select test suites specified command-line
argument "-run". Regular expression to select the methods
of test suites specified command-line argument "-m".
Suite object has assertion methods.
*/
package suite

import (
	"reflect"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"

	"gotest.tools/v3/assert"
	issert "gotest.tools/v3/internal/assert"
)

// TestingSuite used is the interface for a test suite
type TestingSuite interface {
	T() *testing.T
	// SetT is internal, do not use it.
	// Deprecated: SetT should be unexported.
	SetT(*testing.T)
}

// Suite is an implementation of TestingSuite which can be embedded in a test
// suite.
type Suite struct {
	t *testing.T
}

// T retrieves the current *testing.T context.
func (s *Suite) T() *testing.T {
	return s.t
}

// SetT sets the current *testing.T context.
func (s *Suite) SetT(t *testing.T) {
	s.t = t
}

// Run all the tests in a testing suite
func Run(t *testing.T, suite TestingSuite) {
	suite.SetT(t)

	if s, ok := suite.(setupSuite); ok {
		s.SetupSuite()
	}
	if s, ok := suite.(teardownSuite); ok {
		defer s.TearDownSuite()
	}

	suiteType := reflect.TypeOf(suite)
	for index := 0; index < suiteType.NumMethod(); index++ {
		method := suiteType.Method(index)
		if !isTestMethod(method.Name) {
			continue
		}
		t.Run(method.Name, newTestFunc(suite, method))
	}
}

func newTestFunc(suite TestingSuite, method reflect.Method) func(*testing.T) {
	suiteType := reflect.TypeOf(suite)
	return func(t *testing.T) {
		parentT := suite.T()
		suite.SetT(t)
		if s, ok := suite.(setupTest); ok {
			s.SetupTest()
		}
		suiteName := suiteType.Elem().Name()
		if s, ok := suite.(beforeTest); ok {
			s.BeforeTest(suiteName, method.Name)
		}
		defer func() {
			if s, ok := suite.(afterTest); ok {
				s.AfterTest(suiteName, method.Name)
			}
			if s, ok := suite.(teardownTest); ok {
				s.TearDownTest()
			}
			suite.SetT(parentT)
		}()
		method.Func.Call([]reflect.Value{reflect.ValueOf(suite)})
	}
}

func isTestMethod(name string) bool {
	return strings.HasPrefix(name, "Test") && nextRuneIsUpperCase(name[4:])
}

func nextRuneIsUpperCase(r string) bool {
	next, _ := utf8.DecodeRuneInString(r)
	return unicode.IsUpper(next)
}

// Assert performs a comparison, marks the test as having failed if the comparison
// returns false, and stops execution immediately.
//
// This is equivalent to assert.Assert(t, comparison).
func (s *Suite) Assert(comparison assert.BoolOrComparison, msgAndArgs ...interface{}) {
	if ht, ok := testing.TB(s.t).(helperT); ok {
		ht.Helper()
	}

	if !issert.Eval(s.t, issert.ArgsAtZeroIndex, comparison, msgAndArgs...) {
		s.t.FailNow()
	}
}

// Check performs a comparison and marks the test as having failed if the comparison
// returns false. Returns the result of the comparison.
func (s *Suite) Check(comparison assert.BoolOrComparison, msgAndArgs ...interface{}) bool {
	if ht, ok := testing.TB(s.t).(helperT); ok {
		ht.Helper()
	}
	if !issert.Eval(s.t, issert.ArgsAtZeroIndex, comparison, msgAndArgs...) {
		s.t.Fail()
		return false
	}
	return true
}

type setupSuite interface {
	SetupSuite()
}

type setupTest interface {
	SetupTest()
}

type teardownSuite interface {
	TearDownSuite()
}

type teardownTest interface {
	TearDownTest()
}

type beforeTest interface {
	BeforeTest(suiteName, testName string)
}

type afterTest interface {
	AfterTest(suiteName, testName string)
}

type helperT interface {
	Helper()
}
