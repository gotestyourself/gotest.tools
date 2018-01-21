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
	"flag"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"gotest.tools/v3/assert"
)

var methodPattern = &regexpValue{}

func init() {
	flag.Var(methodPattern, "test.m", "only run tests that match the regexp")
}

// TestingSuite used is the interface for a test suite
type TestingSuite interface {
	T() *testing.T
	SetT(*testing.T)
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

// Suite is an implementation of TestingSuite which can be embedded in a test
// suite.
type Suite struct {
	t *testing.T
}

// T retrieves the current *testing.T context.
func (suite *Suite) T() *testing.T {
	return suite.t
}

// SetT sets the current *testing.T context.
func (suite *Suite) SetT(t *testing.T) {
	suite.t = t
}

// Assert performs a comparison, marks the test as having failed if the comparison
// returns false, and stops execution immediately.
//
// This is equivalent to assert.Assert(t, comparison).
func (suite *Suite) Assert(comparison assert.BoolOrComparison, msgAndArgs ...interface{}) {
	if ht, ok := testing.TB(suite.t).(helperT); ok {
		ht.Helper()
	}
	// TODO: will print `comparison` instead of caller ast when used with bool
	assert.Assert(suite.t, comparison, msgAndArgs...)
}

// Check performs a comparison and marks the test as having failed if the comparison
// returns false. Returns the result of the comparison.
func (suite *Suite) Check(comparison assert.BoolOrComparison, msgAndArgs ...interface{}) bool {
	if ht, ok := testing.TB(suite.t).(helperT); ok {
		ht.Helper()
	}
	// TODO: will print `comparison` instead of caller ast when used with bool
	return assert.Check(suite.t, comparison, msgAndArgs...)
}

// Run all the tests in a testing suite
func Run(t *testing.T, suite TestingSuite) {
	suite.SetT(t)

	if s, ok := suite.(setupSuite); ok {
		s.SetupSuite()
	}
	defer func() {
		if s, ok := suite.(teardownSuite); ok {
			s.TearDownSuite()
		}
	}()

	suiteType := reflect.TypeOf(suite)
	tests := []testing.InternalTest{}
	for index := 0; index < suiteType.NumMethod(); index++ {
		method := suiteType.Method(index)
		if !isTestMethod(method.Name) {
			continue
		}
		test := testing.InternalTest{
			Name: method.Name,
			F:    newTestFunc(suite, method),
		}
		tests = append(tests, test)
	}
	runTests(t, tests)
}

func newTestFunc(suite TestingSuite, method reflect.Method) func(*testing.T) {
	suiteType := reflect.TypeOf(suite)
	return func(t *testing.T) {
		parentT := suite.T()
		suite.SetT(t)
		if s, ok := suite.(setupTest); ok {
			s.SetupTest()
		}
		if s, ok := suite.(beforeTest); ok {
			s.BeforeTest(suiteType.Elem().Name(), method.Name)
		}
		defer func() {
			if s, ok := suite.(afterTest); ok {
				s.AfterTest(suiteType.Elem().Name(), method.Name)
			}
			if s, ok := suite.(teardownTest); ok {
				s.TearDownTest()
			}
			suite.SetT(parentT)
		}()
		method.Func.Call([]reflect.Value{reflect.ValueOf(suite)})
	}
}

type runner interface {
	Run(name string, f func(t *testing.T)) bool
}

func runTests(t testing.TB, tests []testing.InternalTest) {
	r, ok := t.(runner)
	if !ok { // backwards compatibility with Go 1.6 and below
		allTestsFilter := func(_, _ string) (bool, error) { return true, nil }
		if !testing.RunTests(allTestsFilter, tests) {
			t.Fail()
		}
		return
	}

	for _, test := range tests {
		r.Run(test.Name, test.F)
	}
}

// TODO: should also check the next character after Test is uppercase
func isTestMethod(name string) bool {
	if !strings.HasPrefix(name, "Test") {
		return false
	}
	return methodPattern.Match(name)
}

type regexpValue struct {
	re *regexp.Regexp
}

func (v *regexpValue) String() string {
	if v.re == nil {
		return ""
	}
	return v.re.String()
}

func (v *regexpValue) Set(value string) error {
	re, err := regexp.Compile(value)
	v.re = re
	return err
}

func (v *regexpValue) Match(value string) bool {
	if v.re == nil {
		return true
	}
	return v.re.MatchString(value)
}
