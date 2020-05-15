package suite

import (
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

type fakeSuite struct {
	Suite

	suiteT  *testing.T
	counter int

	beforeTestCalls []string
	testCalls       []string
	afterTestCalls  []string
}

func (s *fakeSuite) assertAndIncrement(expected int) {
	if ht, ok := testing.TB(s.t).(helperT); ok {
		ht.Helper()
	}
	s.Assert(is.Equal(s.counter, expected))
	s.counter++
}

func (s *fakeSuite) baseCount() int {
	return len(s.afterTestCalls) * 4
}

func (s *fakeSuite) SetupSuite() {
	s.assertAndIncrement(0)
	s.Assert(is.Equal(s.suiteT, s.T()))
}

func (s *fakeSuite) SetupTest() {
	s.assertAndIncrement(s.baseCount() + 1)
	s.Assert(s.suiteT != s.T())
}

func (s *fakeSuite) BeforeTest(suiteName, testName string) {
	s.assertAndIncrement(s.baseCount() + 2)
	s.Assert(is.Equal(suiteName, "fakeSuite"))
	s.beforeTestCalls = append(s.beforeTestCalls, testName)
	s.Assert(s.suiteT != s.T())
}

func (s *fakeSuite) AfterTest(suiteName, testName string) {
	s.assertAndIncrement(s.baseCount() + 3)
	s.Assert(is.Equal(suiteName, "fakeSuite"))
	s.afterTestCalls = append(s.afterTestCalls, testName)
	s.Assert(s.suiteT != s.T())
}

func (s *fakeSuite) TearDownTest() {
	s.assertAndIncrement(s.baseCount())
	s.Assert(s.suiteT != s.T())
}

func (s *fakeSuite) TearDownSuite() {
	s.assertAndIncrement(s.baseCount() + 1)
	s.Assert(is.Equal(s.suiteT, s.T()))
}

func (s *fakeSuite) TestOne() {
	s.testCalls = append(s.testCalls, "TestOne")
	s.Assert(s.suiteT != s.T())
}

func (s *fakeSuite) TestTwo() {
	s.testCalls = append(s.testCalls, "TestTwo")
}

func (s *fakeSuite) TestSkip() {
	s.testCalls = append(s.testCalls, "TestSkip")
	s.T().Skip()
}

func (s *fakeSuite) NonATestMethod() {
}

func TestRunSuite(t *testing.T) {
	fakeSuite := new(fakeSuite)
	fakeSuite.suiteT = t
	Run(t, fakeSuite)

	expectedCount := 14 // setupSuite=1 + teardownSuite=1 + (numTests=3 * numFixtures=4)
	assert.Equal(t, fakeSuite.counter, expectedCount)

	expected := []string{"TestOne", "TestSkip", "TestTwo"}
	assert.Assert(t, is.DeepEqual(expected, fakeSuite.testCalls))
	assert.Assert(t, is.DeepEqual(expected, fakeSuite.afterTestCalls))
	assert.Assert(t, is.DeepEqual(expected, fakeSuite.beforeTestCalls))
}

func TestIsTestMethod(t *testing.T) {
	var testcases = []struct {
		input    string
		expected bool
	}{
		{input: "Test"},
		{input: "Testnotatest"},
		{input: "Test語"},
		{input: "TestI", expected: true},
		{input: "TestIsOne", expected: true},
	}
	for _, tc := range testcases {
		t.Run(tc.input, func(t *testing.T) {
			assert.Equal(t, isTestMethod(tc.input), tc.expected)
		})
	}
}
