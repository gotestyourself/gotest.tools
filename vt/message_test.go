package vt_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gotest.tools/v3/vt"
)

func TestGot(t *testing.T) {
	type testCase struct {
		id         vt.TestID
		fn         func(t *testing.T)
		want       string
		wantPrefix string
	}

	ft := &fakeT{}
	run := func(t *testing.T, tc testCase) {
		defer ft.Reset()
		tc.fn(t)
		if len(ft.args) != 1 {
			t.Fatalf("no result capture")
		}

		if tc.wantPrefix != "" {
			if got := ft.args[0]; !strings.HasPrefix(got.(string), tc.wantPrefix) {
				t.Fatalf("Got(...)\ngot:  %v\nwanted prefix: %v", got, tc.wantPrefix)
			}
			return
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
					ft.Error(vt.Got(err))
				}
			},
			want: `someFunc("arga") returned error: failed to do something (*errors.errorString), wanted ErrorType`,
		},
		{
			id: vt.ID("cmp.Diff"),
			fn: func(t *testing.T) {
				doAThing := func() string {
					return "the actual value"
				}
				want := "the wanted value"
				got := doAThing()

				if diff := cmp.Diff(got, want); diff != "" {
					ft.Fatal(vt.Got(diff))
				}
			},
			wantPrefix: "doAThing() returned a different result (-got +want):\n",
		},
		{
			id: vt.ID("err != nil with comments"),
			fn: func(t *testing.T) {
				if err := someFunc("arga"); err != nil {
					ft.Fatal(vt.Got(err)) // some was not available
				}
			},
			want: `someFunc("arga") returned error: failed to do something
some was not available
`,
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

func (f *fakeT) Error(args ...any) {
	f.args = args
}

func (f *fakeT) Reset() {
	f.args = nil
}

func TestGotWant(t *testing.T) {
	type testCase struct {
		name string
		got  []any
		want string
	}

	run := func(t *testing.T, tc testCase) {

	}

	for _, tc := range []testCase{
		{
			name: "",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			run(t, tc)
		})
	}
}
