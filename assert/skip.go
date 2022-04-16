package assert

import (
	"fmt"
	"path"
	"reflect"
	"runtime"
	"strings"

	"gotest.tools/v3/internal/format"
	"gotest.tools/v3/internal/source"
)

// SkipT is the interface accepted by SkipIf to skip tests. It is implemented by
// testing.T, and testing.B.
type SkipT interface {
	Skip(args ...interface{})
	Log(args ...interface{})
}

// SkipResult may be returned by a function used with SkipIf to provide a
// detailed message to use as part of the skip message.
type SkipResult interface {
	Skip() bool
	Message() string
}

// BoolOrCheckFunc can be a bool, func() bool, or func() SkipResult. Other
// types will panic. See SkipIf for details about how this type is used.
type BoolOrCheckFunc interface{}

// SkipIf skips the test if the condition evaluates to true. If the condition
// evaluates to false then SkipIf does nothing. SkipIf is a convenient way of
// skipping tests and using the literal source of the condition as the text of
// the skip message.
//
// For example, this usage would produce the following skip message:
//
//   assert.SkipIf(t, runtime.GOOS == "windows", "not supported")
//   // filename.go:11: runtime.GOOS == "windows": not supported
//
// The condition argument may be one of the following:
//
//   bool
//     The test will be skipped if the value is true. The literal source of the
//     expression passed to SkipIf will be used as the skip message.
//
//   func() bool
//     The test will be skipped if the function returns true. The name of the
//     function will be used as the skip message.
//
//   func() SkipResult
//     The test will be skipped if SkipResult.Skip return true. Both the name
//     of the function and the return value of SkipResult.Message will be used
//     as the skip message.
//
// Extra details can be added to the skip message using msgAndArgs. msgAndArgs
// may be either a single string, or a format string and args that will be
// passed to fmt.Sprintf.
func SkipIf(t SkipT, condition BoolOrCheckFunc, msgAndArgs ...interface{}) {
	if ht, ok := t.(helperT); ok {
		ht.Helper()
	}
	switch check := condition.(type) {
	case bool:
		ifCondition(t, check, msgAndArgs...)
	case func() bool:
		if check() {
			t.Skip(format.WithCustomMessage(getFunctionName(check), msgAndArgs...))
		}
	case func() SkipResult:
		result := check()
		if result.Skip() {
			msg := getFunctionName(check) + ": " + result.Message()
			t.Skip(format.WithCustomMessage(msg, msgAndArgs...))
		}
	default:
		panic(fmt.Sprintf("invalid type for condition arg: %T", check))
	}
}

func getFunctionName(function interface{}) string {
	funcPath := runtime.FuncForPC(reflect.ValueOf(function).Pointer()).Name()
	return strings.SplitN(path.Base(funcPath), ".", 2)[1]
}

func ifCondition(t SkipT, condition bool, msgAndArgs ...interface{}) {
	if ht, ok := t.(helperT); ok {
		ht.Helper()
	}
	if !condition {
		return
	}
	const (
		stackIndex = 2
		argPos     = 1
	)
	source, err := source.FormattedCallExprArg(stackIndex, argPos)
	if err != nil {
		t.Log(err.Error())
		t.Skip(format.Message(msgAndArgs...))
	}
	t.Skip(format.WithCustomMessage(source, msgAndArgs...))
}
