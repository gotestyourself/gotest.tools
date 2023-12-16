package vt_test

import (
	"errors"
	"fmt"
	"testing"

	"gotest.tools/v3/vt"
)

func TestGot(t *testing.T) {
	type testCase struct {
		id   vt.TestID
		fn   func(t *testing.T)
		want string
	}

	ft := &fakeT{}
	run := func(t *testing.T, tc testCase) {
		defer ft.Reset()
		tc.fn(t)
		if len(ft.args) != 1 {
			t.Fatalf("no result capture")
		}
		if got := ft.args[0]; got != tc.want {
			t.Fatalf("Got(...)\ngot:  %v\nwant: %v", got, tc.want)
		}
	}

	someFunc := func(...any) error {
		return fmt.Errorf("failed to do something")
	}

	testCases := []testCase{
		{
			id: vt.ID("err != nil assigned from function in if block"),
			fn: func(t *testing.T) {
				if err := someFunc("arga"); err != nil {
					ft.Fatal(vt.Got(err))
				}
			},
			want: `someFunc("arga") returned error: failed to do something`,
		},
		{
			id: vt.ID("err != nil assigned from function"),
			fn: func(t *testing.T) {
				err := someFunc("arga")
				if err != nil {
					ft.Fatal(vt.Got(err))
				}
			},
			want: `someFunc("arga") returned error: failed to do something`,
		},
		{
			id: vt.ID("err != nil declared from function"),
			fn: func(t *testing.T) {
				var err = someFunc("arga")
				if err != nil {
					ft.Fatal(vt.Got(err))
				}
			},
			want: `someFunc("arga") returned error: failed to do something`,
		},
		{
			id: vt.ID("errors.Is"),
			fn: func(t *testing.T) {
				var errSentinel = fmt.Errorf("some text")

				err := someFunc("arga")
				if !errors.Is(err, errSentinel) {
					ft.Fatal(vt.Got(err))
				}
			},
			want: `someFunc("arga") returned error: failed to do something, wanted errSentinel`,
		},
		{
			id: vt.ID("errors.As"),
			fn: func(t *testing.T) {
				err := someFunc("arga")
				typedErr := &ErrorType{}
				if !errors.As(err, &typedErr) {
					ft.Fatal(vt.Got(err))
				}
			},
			want: `someFunc("arga") returned error: failed to do something (*errors.errorString), wanted ErrorType`,
		},

		// TODO: cases for assignment from other expr? channel?
		// TODO: cases for err != errSentinel, etc
	}
	for _, tc := range testCases {
		t.Run(tc.id.Name, func(t *testing.T) {
			tc.id.PrintPosition()
			run(t, tc)
		})
	}
}

type ErrorType struct{}

func (e *ErrorType) Error() string {
	return "this type of error"
}

type fakeT struct {
	args []any
}

func (f *fakeT) Fatal(args ...any) {
	f.args = args
}

func (f *fakeT) Reset() {
	f.args = nil
}
