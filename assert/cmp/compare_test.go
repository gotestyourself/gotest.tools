package cmp

import (
	"fmt"
	"testing"
)

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
			expectedMessage: "expected [A b c] to have length 2",
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
			expectedMessage: "expected abcd to have length 3",
		},
	}

	for _, testcase := range testcases {
		t.Run(fmt.Sprintf("%v len=%d", testcase.seq, testcase.length), func(t *testing.T) {
			success, message := Len(testcase.seq, testcase.length)()
			if success != testcase.expectedSuccess {
				t.Errorf("expected success %v, got %v", testcase.expectedSuccess, success)
			}

			if message != testcase.expectedMessage {
				t.Errorf("expected message %q, got %q", testcase.expectedMessage, message)
			}
		})
	}
}
