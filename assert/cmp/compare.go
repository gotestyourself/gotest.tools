/*Package cmp provides Comparisons for Assert and Check*/
package cmp

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/google/go-cmp/cmp"
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
			return false, fmt.Sprintf("expected no error, got %s", lastArg)
		case nil:
			return true, ""
		default:
			return false, fmt.Sprintf(
				"last argument to NoError() must be an error, got %T", lastArg)
		}
	}
}

// TODO: test
func Zero(arg interface{}) func() (bool, string) {
	return func() (bool, string) {
		zero := reflect.Zero(reflect.TypeOf(arg))
		value := reflect.ValueOf(arg)
		return zero == value, fmt.Sprintf("%v is not zero", arg)
	}
}

// TODO: test
func NotZero(arg interface{}) func() (bool, string) {
	return func() (bool, string) {
		zero := reflect.Zero(reflect.TypeOf(arg))
		value := reflect.ValueOf(arg)
		return zero != value, fmt.Sprintf("%v is zero", arg)
	}
}

// TODO: test
// TODO: use == before DeepEqual, check for nils, like ObjectsAreEqual
// TODO: document that reflect.DeepEqual is used
func Contains(seq interface{}, item interface{}) func() (bool, string) {
	return func() (bool, string) {
		list := reflect.ValueOf(seq)
		itemValue := reflect.ValueOf(item)
		msg := fmt.Sprintf("%v does not contains %v", seq, item)

		switch list.Type().Kind() {
		case reflect.String:
			success := strings.Contains(list.String(), itemValue.String())
			return success, msg
		case reflect.Map:
			mapKeys := list.MapKeys()
			for i := 0; i < len(mapKeys); i++ {
				if reflect.DeepEqual(mapKeys[i].Interface(), itemValue) {
					return true, ""
				}
			}
			return false, msg
		case reflect.Slice, reflect.Array, reflect.Chan:
			for i := 0; i < list.Len(); i++ {
				if reflect.DeepEqual(list.Index(i).Interface(), itemValue) {
					return true, ""
				}
			}
			return false, msg
		default:
			return false, fmt.Sprintf("type %T does not contain items", seq)
		}
	}
}

// TODO: test
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
