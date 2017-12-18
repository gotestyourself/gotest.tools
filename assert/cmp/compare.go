/*Package cmp provides Comparisons for Assert and Check*/
package cmp

import (
	"fmt"
	"reflect"

	"github.com/google/go-cmp/cmp"
)

// Compare two complex values using github.com/google/go-cmp/cmp and
// succeeds if the values are equal
func Compare(x, y interface{}, opts ...cmp.Option) func() (success bool, message string) {
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
func Len(seq interface{}, expected int) func() (success bool, message string) {
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
