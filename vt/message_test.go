package vt

import (
	"errors"
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

	someFunc := func(...any) error {
		return fmt.Errorf("failed to do something")
	}

	testCases := []testCase{
		{
			id: ID("err assigned from function in if block"),
			fn: func(t *testing.T) string {
				var got string
				if err := someFunc("arga"); err != nil {
					got = Message(err)
				}
				return got
			},
			want: `someFunc("arga") returned an error: failed to do something`,
		},
		{
			id: ID("err assigned from function"),
			fn: func(t *testing.T) string {
				var got string
				err := someFunc("arga")
				if err != nil {
					got = Message(err)
				}
				return got
			},
			want: `someFunc("arga") returned an error: failed to do something`,
		},
		{
			id: ID("err declared from function"),
			fn: func(t *testing.T) string {
				var got string
				var err = someFunc("arga")
				if err != nil {
					got = Message(err)
				}
				return got
			},
			want: `someFunc("arga") returned an error: failed to do something`,
		},
		{
			id: ID("err incorrectly used without want"),
			fn: func(t *testing.T) string {
				var errSentinel = fmt.Errorf("sentinel")

				var got string
				var err = someFunc("arga")
				if !errors.Is(err, errSentinel) {
					got = Message(err)
				}
				return got
			},
			want: `someFunc("arga") returned an error: failed to do something, wanted errSentinel`,
		},

		// TODO: cases for assignment from other expr? channel?
		// TODO: cases for incorrect usage: errors.{Is,As}, err != errSentinel, etc
	}
	for _, tc := range testCases {
		t.Run(tc.id.Name, func(t *testing.T) {
			tc.id.PrintPosition()
			run(t, tc)
		})
	}
}
