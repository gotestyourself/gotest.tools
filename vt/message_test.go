package vt

import (
	"fmt"
	"testing"
)

func TestMessage(t *testing.T) {
	type testCase struct {
		id   TestID
		fn   func(t *testing.T) string
		want string
	}

	run := func(t *testing.T, tc testCase) {
		got := tc.fn(t)
		if got != tc.want {
			t.Fatalf("Message(...)\ngot:  %v\nwant: %v", got, tc.want)
		}
	}

	errFunc := func(...any) error {
		return fmt.Errorf("failed to do something")
	}

	testCases := []testCase{
		{
			id: ID("err assigned from function"),
			fn: func(t *testing.T) string {
				var got string
				if err := errFunc("a", 1, nil); err != nil {
					got = Message(err)
				}
				return got
			},
			want: `errFunc("a", 1, nil) returned an error: failed to do something`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.id.Name, func(t *testing.T) {
			tc.id.PrintPosition()
			run(t, tc)
		})
	}
}
