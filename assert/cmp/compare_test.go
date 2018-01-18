package cmp

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"testing"

	"github.com/pkg/errors"
)

func TestCompare(t *testing.T) {
	success, msg := Compare([]string{"a", "b"}, []string{"b", "a"})()
	assertFailure(t, success, msg, `
{[]string}:
	-: []string{"a", "b"}
	+: []string{"b", "a"}
`)

	success, msg = Compare([]string{"a"}, []string{"a"})()
	assertSuccess(t, success, msg)
}

type Stub struct {
	unx int
}

func TestCompareWithUnexported(t *testing.T) {
	success, msg := Compare(Stub{}, Stub{unx: 1})()
	assertFailure(t, success, msg, `cannot handle unexported field: {cmp.Stub}.unx
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
			success, message := Len(testcase.seq, testcase.length)()
			if testcase.expectedSuccess {
				assertSuccess(t, success, message)
			} else {
				assertFailure(t, success, message, testcase.expectedMessage)
			}
		})
	}
}

type stubError struct{}

func (e *stubError) Error() string { return "stub error" }

func TestNilError(t *testing.T) {
	var s *stubError
	success, message := NilError(s)()
	assertSuccess(t, success, message)

	success, message = NilError(nil)()
	assertSuccess(t, success, message)

	var e error
	success, message = NilError(e)()
	assertSuccess(t, success, message)

	buf := new(bytes.Buffer)
	success, message = NilError(buf.WriteString("ok"))()
	assertSuccess(t, success, message)

	s = &stubError{}
	success, message = NilError(s)()
	assertFailure(t, success, message, "error is not nil: stub error")

	e = &stubError{}
	success, message = NilError(e)()
	assertFailure(t, success, message, "error is not nil: stub error")
}

func TestPanics(t *testing.T) {
	panicker := func() {
		panic("AHHHHHHHHHHH")
	}

	success, message := Panics(panicker)()
	assertSuccess(t, success, message)

	success, message = Panics(func() {})()
	assertFailure(t, success, message, "did not panic")
}

type innerstub struct {
	num int
}

type stub struct {
	stub innerstub
	num  int
}

func TestDeepEqual(t *testing.T) {
	var testcases = []struct {
		left     interface{}
		right    interface{}
		expected bool
	}{
		{nil, nil, true},
		{7, 7, true},
		{false, false, true},
		{stub{innerstub{1}, 2}, stub{innerstub{1}, 2}, true},
		{[]int{1, 2, 3}, []int{1, 2, 3}, true},
		{[]byte(nil), []byte(nil), true},
		{nil, []byte(nil), false},
		{1, uint64(1), false},
		{7, "7", false},
	}
	for _, testcase := range testcases {
		if reflect.DeepEqual(testcase.left, testcase.right) != testcase.expected {
			t.Errorf("deepEqual(%v, %v) did not return %v",
				testcase.left, testcase.right, testcase.expected)
		}
	}
}

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
			success, message := Contains(testcase.seq, testcase.item)()
			if testcase.expected {
				assertSuccess(t, success, message)
			} else {
				assertFailure(t, success, message, testcase.expectedMsg)
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

	success, msg := EqualMultiLine(left, right)()
	assertFailure(t, success, msg, expected)
}

func TestError(t *testing.T) {
	success, message := Error(nil, "the error message")()
	assertFailure(t, success, message, "expected an error, got nil")

	success, message = Error(errors.New("other"), "the error message")()
	assertFailure(t, success, message,
		`expected error message "the error message", got "other"`)

	msg := "the message"
	success, message = Error(errors.New(msg), msg)()
	assertSuccess(t, success, message)
}

func TestErrorContains(t *testing.T) {
	success, message := ErrorContains(nil, "the error message")()
	assertFailure(t, success, message, "expected an error, got nil")

	success, message = ErrorContains(errors.New("other"), "the error")()
	assertFailure(t, success, message,
		`expected error message to contain "the error", got "other"`)

	msg := "the full message"
	success, message = ErrorContains(errors.New(msg), "full")()
	assertSuccess(t, success, message)
}

func TestNil(t *testing.T) {
	success, message := Nil(nil)()
	assertSuccess(t, success, message)

	var s *string
	success, message = Nil(s)()
	assertSuccess(t, success, message)

	var closer io.Closer
	success, message = Nil(closer)()
	assertSuccess(t, success, message)

	success, message = Nil("wrong")()
	assertFailure(t, success, message, "wrong (type string) can not be nil")

	notnil := "notnil"
	success, message = Nil(&notnil)()
	assertFailure(t, success, message, "notnil (type *string) is not nil")

	success, message = Nil([]string{"a"})()
	assertFailure(t, success, message, "[a] (type []string) is not nil")
}

type testingT interface {
	Errorf(msg string, args ...interface{})
}

type helperT interface {
	Helper()
}

func assertSuccess(t testingT, success bool, message string) {
	if ht, ok := t.(helperT); ok {
		ht.Helper()
	}
	if !success {
		t.Errorf("expected success, but got failure with message %q", message)
	}
}

func assertFailure(t testingT, success bool, message string, expected string) {
	if ht, ok := t.(helperT); ok {
		ht.Helper()
	}
	if success {
		t.Errorf("expected failure")
	}
	if message != expected {
		t.Errorf("expected \n%q\ngot\n%q\n", expected, message)
	}
}
