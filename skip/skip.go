/*Package skip provides functions for skipping based on a condition.
 */
package skip

import (
	"fmt"
	"path"
	"reflect"
	"runtime"
	"strings"

	"github.com/gotestyourself/gotestyourself/internal/source"
)

type skipT interface {
	Skip(args ...interface{})
	Log(args ...interface{})
}

// BoolOrCheckFunc can be a bool or func() bool, other types will panic
type BoolOrCheckFunc interface{}

// If skips the test if the check function returns true. The skip message will
// contain the name of the check function. Extra message text can be passed as a
// format string with args
func If(t skipT, condition BoolOrCheckFunc, msgAndArgs ...interface{}) {
	switch check := condition.(type) {
	case bool:
		ifCondition(t, check, msgAndArgs...)
	case func() bool:
		if check() {
			t.Skip(formatWithCustomMessage(
				getFunctionName(check),
				formatMessage(msgAndArgs...)))
		}
	default:
		panic(fmt.Sprintf("invalid type for condition arg: %T", check))
	}
}

func getFunctionName(function func() bool) string {
	funcPath := runtime.FuncForPC(reflect.ValueOf(function).Pointer()).Name()
	return strings.SplitN(path.Base(funcPath), ".", 2)[1]
}

// IfCondition skips the test if the condition is true. The skip message will
// contain the source of the expression passed as the condition. Extra message
// text can be passed as a format string with args.
//
// Deprecated: Use If() which now accepts bool arguments
func IfCondition(t skipT, condition bool, msgAndArgs ...interface{}) {
	ifCondition(t, condition, msgAndArgs...)
}

func ifCondition(t skipT, condition bool, msgAndArgs ...interface{}) {
	if !condition {
		return
	}
	const (
		stackIndex = 4
		argPos     = 1
	)
	source, err := source.GetCondition(stackIndex, argPos)
	if err != nil {
		t.Log(err.Error())
		t.Skip(formatMessage(msgAndArgs...))
	}
	t.Skip(formatWithCustomMessage(source, formatMessage(msgAndArgs...)))
}

func formatMessage(msgAndArgs ...interface{}) string {
	switch len(msgAndArgs) {
	case 0:
		return ""
	case 1:
		return msgAndArgs[0].(string)
	default:
		return fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
	}
}

func formatWithCustomMessage(source, custom string) string {
	switch {
	case custom == "":
		return source
	case source == "":
		return custom
	}
	return fmt.Sprintf("%s: %s", source, custom)
}
