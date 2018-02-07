/*Package cmp provides Comparisons for Assert and Check*/
package cmp

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/pmezard/go-difflib/difflib"
)

// Comparison is a function which compares values and returns ResultSuccess if
// the actual value matches the expected value. If the values do not match the
// Result will contain a message about why it failed.
type Comparison func() Result

// DeepEqual compares two values using https://godoc.org/github.com/google/go-cmp/cmp
// and succeeds if the values are equal.
//
// The comparison can be customized using comparison Options.
func DeepEqual(x, y interface{}, opts ...cmp.Option) Comparison {
	return func() (result Result) {
		defer func() {
			if panicmsg, handled := handleCmpPanic(recover()); handled {
				result = ResultFailure(panicmsg)
			}
		}()
		diff := cmp.Diff(x, y, opts...)
		return toResult(diff == "", "\n"+diff)
	}
}

func handleCmpPanic(r interface{}) (string, bool) {
	if r == nil {
		return "", false
	}
	panicmsg, ok := r.(string)
	if !ok {
		panic(r)
	}
	switch {
	case strings.HasPrefix(panicmsg, "cannot handle unexported field"):
		return panicmsg, true
	}
	panic(r)
}

func toResult(success bool, msg string) Result {
	if success {
		return ResultSuccess
	}
	return ResultFailure(msg)
}

// Equal succeeds if x == y.
func Equal(x, y interface{}) Comparison {
	return func() Result {
		switch {
		case x == y:
			return ResultSuccess
		case isMultiLineStringCompare(x, y):
			return multiLineStringDiffResult(x.(string), y.(string))
		}
		return ResultFailureTemplate(`
			{{- .Data.x}} (
				{{- with callArg 0 }}{{ formatNode . }} {{end -}}
				{{- printf "%T" .Data.x -}}
			) != {{ .Data.y}} (
				{{- with callArg 1 }}{{ formatNode . }} {{end -}}
				{{- printf "%T" .Data.y -}}
			)`,
			map[string]interface{}{"x": x, "y": y})
	}
}

func isMultiLineStringCompare(x, y interface{}) bool {
	strX, ok := x.(string)
	if !ok {
		return false
	}
	strY, ok := y.(string)
	if !ok {
		return false
	}
	return strings.Contains(strX, "\n") || strings.Contains(strY, "\n")
}

func multiLineStringDiffResult(x, y string) Result {
	diff, err := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
		A:       difflib.SplitLines(x),
		B:       difflib.SplitLines(y),
		Context: 3,
	})
	if err != nil {
		return ResultFailure(fmt.Sprintf("failed to diff: %s", err))
	}
	return ResultFailureTemplate(`
--- {{ with callArg 0 }}{{ formatNode . }}{{else}}←{{end}}
+++ {{ with callArg 1 }}{{ formatNode . }}{{else}}→{{end}}
{{ .Data.diff }}`,
		map[string]interface{}{"diff": diff})
}

// Len succeeds if the sequence has the expected length.
func Len(seq interface{}, expected int) Comparison {
	return func() (result Result) {
		defer func() {
			if e := recover(); e != nil {
				result = ResultFailure(fmt.Sprintf("type %T does not have a length", seq))
			}
		}()
		value := reflect.ValueOf(seq)
		length := value.Len()
		if length == expected {
			return ResultSuccess
		}
		msg := fmt.Sprintf("expected %s (length %d) to have length %d", seq, length, expected)
		return ResultFailure(msg)
	}
}

// Contains succeeds if item is in collection. Collection may be a string, map,
// slice, or array.
//
// If collection is a string, item must also be a string, and is compared using
// strings.Contains().
// If collection is a Map, contains will succeed if item is a key in the map.
// If collection is a slice or array, item is compared to each item in the
// sequence using reflect.DeepEqual().
func Contains(collection interface{}, item interface{}) Comparison {
	return func() Result {
		colValue := reflect.ValueOf(collection)
		if !colValue.IsValid() {
			return ResultFailure(fmt.Sprintf("nil does not contain items"))
		}
		msg := fmt.Sprintf("%v does not contain %v", collection, item)

		itemValue := reflect.ValueOf(item)
		switch colValue.Type().Kind() {
		case reflect.String:
			if itemValue.Type().Kind() != reflect.String {
				return ResultFailure("string may only contain strings")
			}
			return toResult(
				strings.Contains(colValue.String(), itemValue.String()),
				fmt.Sprintf("string %q does not contain %q", collection, item))

		case reflect.Map:
			if itemValue.Type() != colValue.Type().Key() {
				return ResultFailure(fmt.Sprintf(
					"%v can not contain a %v key", colValue.Type(), itemValue.Type()))
			}
			return toResult(colValue.MapIndex(itemValue).IsValid(), msg)

		case reflect.Slice, reflect.Array:
			for i := 0; i < colValue.Len(); i++ {
				if reflect.DeepEqual(colValue.Index(i).Interface(), item) {
					return ResultSuccess
				}
			}
			return ResultFailure(msg)
		default:
			return ResultFailure(fmt.Sprintf("type %T does not contain items", collection))
		}
	}
}

// Panics succeeds if f() panics.
func Panics(f func()) Comparison {
	return func() (result Result) {
		defer func() {
			if err := recover(); err != nil {
				result = ResultSuccess
			}
		}()
		f()
		return ResultFailure("did not panic")
	}
}

// Error succeeds if err is a non-nil error, and the error message equals the
// expected message.
func Error(err error, message string) Comparison {
	return func() Result {
		switch {
		case err == nil:
			return ResultFailure("expected an error, got nil")
		case err.Error() != message:
			return ResultFailure(fmt.Sprintf(
				"expected error %q, got %+v", message, err))
		}
		return ResultSuccess
	}
}

// ErrorContains succeeds if err is a non-nil error, and the error message contains
// the expected substring.
func ErrorContains(err error, substring string) Comparison {
	return func() Result {
		switch {
		case err == nil:
			return ResultFailure("expected an error, got nil")
		case !strings.Contains(err.Error(), substring):
			return ResultFailure(fmt.Sprintf(
				"expected error to contain %q, got %+v", substring, err))
		}
		return ResultSuccess
	}
}

// Nil succeeds if obj is a nil interface, pointer, or function.
//
// Use assert.Assert() for comparing errors. Use Len(obj, 0) for comparing slices,
// maps, and channels.
func Nil(obj interface{}) Comparison {
	return func() Result {
		if obj == nil {
			return ResultSuccess
		}
		value := reflect.ValueOf(obj)
		kind := value.Type().Kind()
		if kind >= reflect.Chan && kind <= reflect.Slice {
			if value.IsNil() {
				return ResultSuccess
			}

			return ResultFailure(fmt.Sprintf(
				"%v (type %s) is not nil",
				reflect.Indirect(value),
				value.Type()))
		}

		return ResultFailure(fmt.Sprintf("%v (type %s) can not be nil", value, value.Type()))
	}
}
