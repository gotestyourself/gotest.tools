package cmp

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
)

func TestDeepEqual(t *testing.T) {
	result := DeepEqual([]string{"a", "b"}, []string{"b", "a"})()
	assertFailure(t, result, `
{[]string}:
	-: []string{"a", "b"}
	+: []string{"b", "a"}
`)

	result = DeepEqual([]string{"a"}, []string{"a"})()
	assertSuccess(t, result)
}

type Stub struct {
	unx int
}

func TestDeepEqualeWithUnexported(t *testing.T) {
	result := DeepEqual(Stub{}, Stub{unx: 1})()
	assertFailure(t, result, `cannot handle unexported field: {cmp.Stub}.unx
consider using AllowUnexported or cmpopts.IgnoreUnexported`)
}

func TestLen(t *testing.T) {
	var testcases = []struct {
		seq             interface{}
		length          int
		expectedSuccess bool
		expectedMessage string
	}{
		{
			seq:             []string{"A", "b", "c"},
			length:          3,
			expectedSuccess: true,
		},
		{
			seq:             []string{"A", "b", "c"},
			length:          2,
			expectedMessage: "expected [A b c] (length 3) to have length 2",
		},
		{
			seq:             map[string]int{"a": 1, "b": 2},
			length:          2,
			expectedSuccess: true,
		},
		{
			seq:             [3]string{"a", "b", "c"},
			length:          3,
			expectedSuccess: true,
		},
		{
			seq:             "abcd",
			length:          4,
			expectedSuccess: true,
		},
		{
			seq:             "abcd",
			length:          3,
			expectedMessage: "expected abcd (length 4) to have length 3",
		},
	}

	for _, testcase := range testcases {
		t.Run(fmt.Sprintf("%v len=%d", testcase.seq, testcase.length), func(t *testing.T) {
			result := Len(testcase.seq, testcase.length)()
			if testcase.expectedSuccess {
				assertSuccess(t, result)
			} else {
				assertFailure(t, result, testcase.expectedMessage)
			}
		})
	}
}

type stubError struct{}

func (e *stubError) Error() string { return "stub error" }

func TestNilError(t *testing.T) {
	var s *stubError
	result := NilError(s)()
	assertSuccess(t, result)

	result = NilError(nil)()
	assertSuccess(t, result)

	var e error
	result = NilError(e)()
	assertSuccess(t, result)

	buf := new(bytes.Buffer)
	result = NilError(buf.WriteString("ok"))()
	assertSuccess(t, result)

	s = &stubError{}
	result = NilError(s)()
	assertFailure(t, result, "error is not nil: stub error")

	e = &stubError{}
	result = NilError(e)()
	assertFailure(t, result, "error is not nil: stub error")
}

func TestPanics(t *testing.T) {
	panicker := func() {
		panic("AHHHHHHHHHHH")
	}

	result := Panics(panicker)()
	assertSuccess(t, result)

	result = Panics(func() {})()
	assertFailure(t, result, "did not panic")
}

type innerstub struct {
	num int
}

type stub struct {
	stub innerstub
	num  int
}

func TestDeepEqualEquivalenceToReflectDeepEqual(t *testing.T) {
	var testcases = []struct {
		left  interface{}
		right interface{}
	}{
		{nil, nil},
		{7, 7},
		{false, false},
		{stub{innerstub{1}, 2}, stub{innerstub{1}, 2}},
		{[]int{1, 2, 3}, []int{1, 2, 3}},
		{[]byte(nil), []byte(nil)},
		{nil, []byte(nil)},
		{1, uint64(1)},
		{7, "7"},
	}
	for _, testcase := range testcases {
		expected := reflect.DeepEqual(testcase.left, testcase.right)
		res := DeepEqual(testcase.left, testcase.right, cmpStub)()
		if res.Success() != expected {
			msg := res.(result).FailureMessage()
			t.Errorf("deepEqual(%v, %v) did not return %v (message %s)",
				testcase.left, testcase.right, expected, msg)
		}
	}
}

var cmpStub = cmp.AllowUnexported(stub{}, innerstub{})

func TestContains(t *testing.T) {
	var testcases = []struct {
		seq         interface{}
		item        interface{}
		expected    bool
		expectedMsg string
	}{
		{
			seq:         error(nil),
			item:        0,
			expectedMsg: "nil does not contain items",
		},
		{
			seq:      "abcdef",
			item:     "cde",
			expected: true,
		},
		{
			seq:         "abcdef",
			item:        "foo",
			expectedMsg: `string "abcdef" does not contain "foo"`,
		},
		{
			seq:         "abcdef",
			item:        3,
			expectedMsg: `string may only contain strings`,
		},
		{
			seq:      map[rune]int{'a': 1, 'b': 2},
			item:     'b',
			expected: true,
		},
		{
			seq:         map[rune]int{'a': 1},
			item:        'c',
			expectedMsg: "map[97:1] does not contain 99",
		},
		{
			seq:         map[int]int{'a': 1, 'b': 2},
			item:        'b',
			expectedMsg: "map[int]int can not contain a int32 key",
		},
		{
			seq:      []interface{}{"a", 1, 'a', 1.0, true},
			item:     'a',
			expected: true,
		},
		{
			seq:         []interface{}{"a", 1, 'a', 1.0, true},
			item:        3,
			expectedMsg: "[a 1 97 1 true] does not contain 3",
		},
		{
			seq:      [3]byte{99, 10, 100},
			item:     byte(99),
			expected: true,
		},
		{
			seq:         [3]byte{99, 10, 100},
			item:        byte(98),
			expectedMsg: "[99 10 100] does not contain 98",
		},
	}
	for _, testcase := range testcases {
		name := fmt.Sprintf("%v in %v", testcase.item, testcase.seq)
		t.Run(name, func(t *testing.T) {
			result := Contains(testcase.seq, testcase.item)()
			if testcase.expected {
				assertSuccess(t, result)
			} else {
				assertFailure(t, result, testcase.expectedMsg)
			}
		})
	}
}

func TestEqualMultiLine(t *testing.T) {
	left := `abcd
1234
aaaa
bbbb`

	right := `abcd
1111
aaaa
bbbb`

	expected := `
--- left
+++ right
@@ -1,4 +1,4 @@
 abcd
-1234
+1111
 aaaa
 bbbb
`

	result := EqualMultiLine(left, right)()
	assertFailure(t, result, expected)
}

func TestError(t *testing.T) {
	result := Error(nil, "the error message")()
	assertFailure(t, result, "expected an error, got nil")

	result = Error(errors.New("other"), "the error message")()
	assertFailureHasPrefix(t, result,
		`expected error "the error message", got other`)

	msg := "the message"
	result = Error(errors.New(msg), msg)()
	assertSuccess(t, result)
}

func TestErrorContains(t *testing.T) {
	result := ErrorContains(nil, "the error message")()
	assertFailure(t, result, "expected an error, got nil")

	result = ErrorContains(errors.New("other"), "the error")()
	assertFailureHasPrefix(t, result,
		`expected error to contain "the error", got other`)

	msg := "the full message"
	result = ErrorContains(errors.New(msg), "full")()
	assertSuccess(t, result)
}

func TestNil(t *testing.T) {
	result := Nil(nil)()
	assertSuccess(t, result)

	var s *string
	result = Nil(s)()
	assertSuccess(t, result)

	var closer io.Closer
	result = Nil(closer)()
	assertSuccess(t, result)

	result = Nil("wrong")()
	assertFailure(t, result, "wrong (type string) can not be nil")

	notnil := "notnil"
	result = Nil(&notnil)()
	assertFailure(t, result, "notnil (type *string) is not nil")

	result = Nil([]string{"a"})()
	assertFailure(t, result, "[a] (type []string) is not nil")
}

type testingT interface {
	Errorf(msg string, args ...interface{})
}

type helperT interface {
	Helper()
}

func assertSuccess(t testingT, res Result) {
	if ht, ok := t.(helperT); ok {
		ht.Helper()
	}
	if !res.Success() {
		msg := res.(result).FailureMessage()
		t.Errorf("expected success, but got failure with message %q", msg)
	}
}

func assertFailure(t testingT, res Result, expected string) {
	if ht, ok := t.(helperT); ok {
		ht.Helper()
	}
	if res.Success() {
		t.Errorf("expected failure")
	}
	message := res.(result).FailureMessage()
	if message != expected {
		t.Errorf("expected \n%q\ngot\n%q\n", expected, message)
	}
}

func assertFailureHasPrefix(t testingT, res Result, prefix string) {
	if ht, ok := t.(helperT); ok {
		ht.Helper()
	}
	if res.Success() {
		t.Errorf("expected failure")
	}
	message := res.(result).FailureMessage()
	if !strings.HasPrefix(message, prefix) {
		t.Errorf("expected \n%v\nto start with\n%v\n", message, prefix)
	}
}
