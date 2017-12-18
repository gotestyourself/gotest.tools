/*Package cmp provides Comparisons for Assert and Check*/
package cmp

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/pmezard/go-difflib/difflib"
)

// Compare two complex values using github.com/google/go-cmp/cmp and
// succeeds if the values are equal
func Compare(x, y interface{}, opts ...cmp.Option) func() (bool, string) {
	return func() (bool, string) {
		diff := cmp.Diff(x, y, opts...)
		// TODO: wrap error message?
		return diff == "", diff
	}
}

// Equal compares two values using the == operator
func Equal(x, y interface{}) func() (success bool, message string) {
	return func() (bool, string) {
		return x == y, fmt.Sprintf("%v (%T) != %v (%T)", x, x, y, y)
	}
}

// Len succeeds if the sequence has the expected length
func Len(seq interface{}, expected int) func() (bool, string) {
	return func() (success bool, message string) {
		defer func() {
			if e := recover(); e != nil {
				success = false
				message = fmt.Sprintf("type %T does not have a length", seq)
			}
		}()
		value := reflect.ValueOf(seq)
		if value.Len() == expected {
			return true, ""
		}
		return false, fmt.Sprintf("expected %s to have length %d", seq, expected)
	}
}

// NoError succeeds if the last argument is a nil error
func NoError(args ...interface{}) func() (bool, string) {
	return func() (bool, string) {
		if len(args) == 0 {
			return true, ""
		}
		switch lastArg := args[len(args)-1].(type) {
		case error:
			return false, fmt.Sprintf("expected no error, got %+v", lastArg)
		case nil:
			return true, ""
		default:
			return false, fmt.Sprintf(
				"last argument to NoError() must be an error, got %T", lastArg)
		}
	}
}

// Contains succeeds if item is in the seq. seq may be a string, map, slice, or
// array. reflect.DeepEqual() is used to compare the item with each value in the
// sequence.
func Contains(seq interface{}, item interface{}) func() (bool, string) {
	return func() (bool, string) {
		seqValue := reflect.ValueOf(seq)
		if !seqValue.IsValid() {
			return false, fmt.Sprintf("nil does not contain items")
		}
		msg := fmt.Sprintf("%v does not contain %v", seq, item)

		switch seqValue.Type().Kind() {
		case reflect.String:
			itemValue := reflect.ValueOf(item)
			success := strings.Contains(seqValue.String(), itemValue.String())
			return success, msg
		case reflect.Map:
			mapKeys := seqValue.MapKeys()
			for i := 0; i < len(mapKeys); i++ {
				if reflect.DeepEqual(mapKeys[i].Interface(), item) {
					return true, ""
				}
			}
			return false, msg
		case reflect.Slice, reflect.Array:
			for i := 0; i < seqValue.Len(); i++ {
				if reflect.DeepEqual(seqValue.Index(i).Interface(), item) {
					return true, ""
				}
			}
			return false, msg
		default:
			return false, fmt.Sprintf("type %T does not contain items", seq)
		}
	}
}

// Panics succeeds if f() panics
func Panics(f func()) func() (bool, string) {
	return func() (success bool, message string) {
		defer func() {
			if err := recover(); err != nil {
				success = true
			}
		}()
		f()
		return false, "did not panic"
	}
}

// EqualMultiLine succeeds if the two string are equal. If they are not equal
// the failure message with be a unified diff of the difference.
func EqualMultiLine(x, y string) func() (bool, string) {
	return func() (bool, string) {
		if x == y {
			return true, ""
		}

		diff, err := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
			A:        difflib.SplitLines(x),
			B:        difflib.SplitLines(y),
			FromFile: "left",
			ToFile:   "right",
			Context:  3,
		})
		if err != nil {
			return false, fmt.Sprintf("failed to produce diff: %s", err)
		}
		return false, diff
	}
}

// Error succeeds if err is a non-nil error, and the error message equals the
// expected message.
func Error(err error, message string) func() (bool, string) {
	return func() (bool, string) {
		switch {
		case err == nil:
			return false, "expected an error, got nil"
		case err.Error() != message:
			return false, fmt.Sprintf(
				"expected error message %q, got %q", message, err.Error())
		}
		return true, ""
	}
}

// ErrorContains succeeds if err is a non-nil error, and the error message contains
// the expected substring.
func ErrorContains(err error, substring string) func() (bool, string) {
	return func() (bool, string) {
		switch {
		case err == nil:
			return false, "expected an error, got nil"
		case !strings.Contains(err.Error(), substring):
			return false, fmt.Sprintf(
				"expected error message to contain %q, got %q", substring, err.Error())
		}
		return true, ""
	}
}
