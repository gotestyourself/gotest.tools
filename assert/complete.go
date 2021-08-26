package assert

import (
	"fmt"
	"math/rand"
	"reflect"
	"strings"
)

// Input are the settings used by Complete.
type Input[T any] struct {
	// Original must be set to a function that returns a copy to a pointer of
	// the struct accepted by the complete operation. Original will be called
	// for each field in the struct, and passed as the first argument to Operation.
	Original func() *T

	// Operation accepts two values of the same type and returns true if the
	// operation was successful, and false otherwise.
	// Common operations that may be tested using Complete include:
	//   * equal
	//   * empty
	//   * round tripping between two transformations
	//   * building a hash or map key
	Operation func(original T, modified T) bool
}

// Complete tests that the operation defined by input considers all the fields
// in the type T. T must be a struct.
func Complete[T any](t FatalF, input Input[T]) {
	if th, ok := t.(helperT); ok {
		th.Helper()
	}
	origValue := func() reflect.Value {
		return reflect.Indirect(reflect.ValueOf(input.Original()))
	}
	orig := origValue()

	cfg := config{
		origFn: origValue,
		op: func(modified reflect.Value) bool {
			opFn := reflect.ValueOf(input.Operation)
			return opFn.Call([]reflect.Value{orig, modified})[0].Bool()
		},
	}
	pos := position{
		structType:       orig.Type(),
		getValueForField: func(v reflect.Value) reflect.Value { return v },
	}
	traverseStruct(t, cfg, pos)
}

type FatalF interface {
	Fatalf(format string, args ...interface{})
}

type config struct {
	origFn func() reflect.Value
	op     func(v reflect.Value) bool
}

type position struct {
	structType       reflect.Type
	path             string
	getValueForField lookup
}

func (p position) fieldName(i int) string {
	return p.path + p.structType.Field(i).Name
}

func traverseStruct(t FatalF, cfg config, pos position) {
	if th, ok := t.(helperT); ok {
		th.Helper()
	}
	for i := 0; i < pos.structType.NumField(); i++ {
		sample := cfg.origFn()
		field := pos.getValueForField(sample).Field(i)

		switch f := reflect.Indirect(field); f.Kind() {
		case reflect.Struct:
			// TODO: limit max recurse

			nextPos := position{
				path:             pos.fieldName(i) + ".",
				structType:       field.Type(),
				getValueForField: getFieldFn(pos.getValueForField, i),
			}
			traverseStruct(t, cfg, nextPos)

			// TODO: recurse for slice/array/map
		default:
			fillValue(field.Addr())
			if !cfg.op(sample) {
				t.Fatalf("not complete: field %v is not included", pos.fieldName(i))
			}
		}
	}
}

type lookup func(next reflect.Value) reflect.Value

func getFieldFn(base lookup, index int) lookup {
	return func(next reflect.Value) reflect.Value {
		return base(next).Field(index)
	}
}

func fillValue(addr reflect.Value) {
	if addr.Kind() != reflect.Ptr {
		panic("must be a pointer")
	}
	v := addr.Elem()

	if !v.CanSet() || v.Kind() == reflect.Invalid {
		panic(fmt.Sprintf("%v (%v) is not settable", v, v.Type()))
	}

	// TODO: support some way of setting the seed

	switch v.Kind() {
	case reflect.Bool:
		v.SetBool(!v.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		for {
			next := int64(rand.Int31n(120) + 1)
			if next != v.Int() {
				v.SetInt(next)
				return
			}
		}

	case reflect.Float32, reflect.Float64:
		for {
			next := float64(rand.Float32() + 1)
			if next != v.Float() {
				v.SetFloat(next)
				return
			}
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		// TODO: new value
		v.SetUint(uint64(rand.Uint32() + 1))
	case reflect.String:
		v.SetString(randString())
	case reflect.Slice, reflect.Array:
		// TODO:
		panic("TODO: support slice and array")
	case reflect.Map:
		// TODO:
		panic("TODO: support map")
	case reflect.Interface:
		// TODO:
		panic("TODO: support interface")
	case reflect.Struct:
		panic("structs should be filled by individual field")
	case reflect.Ptr:
		fallthrough
	case reflect.Complex64, reflect.Complex128:
		fallthrough
	case reflect.Chan, reflect.Func, reflect.UnsafePointer:
		panic(fmt.Sprintf("fill: not implemented for kind %v", v.Kind()))
	}
}

// TODO: unicode
func randString() string {
	chars := "abcdefghijklmnopqrstuvwxyz"
	var out strings.Builder
	for i := 0; i <= rand.Intn(20)+5; i++ {
		out.WriteByte(chars[rand.Intn(len(chars))])
	}
	return out.String()
}
