package assert

import (
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"time"
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

	// IgnoreFields is a list of struct field paths that should be skipped. These
	// fields are intentionally not part of the operation. Each value in the list
	// is the path to a field, using dotted notation for nested fields.
	//
	// The example below demonstrates the value that would be used to ignore
	// each of the fields on the struct.
	//
	//   type Request struct {
	//       URL string            // "URL"
	//       Meta struct {         // "Meta"
	//           Label string      // "Meta.Label"
	//       }
	//   }
	IgnoreFields []string

	// Seed is the value used to initialize the random source. Defaults to
	// time.Time.UnixNano of the current time if unset. If a failure is only
	// reproducing with a specific seed, you can set this value to reproduce
	// the failure.
	Seed int64
}

// Complete tests that the operation defined by input considers all the fields
// in the type T. T must be a struct.
func Complete[T any](t FatalF, input Input[T]) {
	if th, ok := t.(helperT); ok {
		th.Helper()
	}
	if input.Seed == 0 {
		input.Seed = time.Now().UnixNano()
	}
	t.Log("using random seed ", input.Seed)

	origFn := func() reflect.Value {
		return reflect.Indirect(reflect.ValueOf(input.Original()))
	}
	orig := origFn()
	cfg := config{
		rand:   rand.New(rand.NewSource(input.Seed)),
		origFn: origFn,
		op: func(modified reflect.Value) bool {
			opFn := reflect.ValueOf(input.Operation)
			return opFn.Call([]reflect.Value{orig, modified})[0].Bool()
		},
		ignored: make(map[string]struct{}, len(input.IgnoreFields)),
	}
	for _, k := range input.IgnoreFields {
		cfg.ignored[k] = struct{}{}
	}
	pos := position{
		structType: orig.Type(),
		getValue:   func(v reflect.Value) reflect.Value { return v },
	}
	traverseStruct(t, cfg, pos)
}

type FatalF interface {
	Log(args ...interface{})
	Fatalf(format string, args ...interface{})
}

// config is the internal version of Input that is used by traverseStruct
type config struct {
	origFn  func() reflect.Value
	op      func(v reflect.Value) bool
	ignored map[string]struct{}
	rand    *rand.Rand
}

// position identifies the position in struct traversal.
type position struct {
	structType reflect.Type
	path       string
	getValue   func(modified reflect.Value) reflect.Value
}

func (p position) fieldName(i int) string {
	return p.path + p.structType.Field(i).Name
}

func traverseStruct(t FatalF, cfg config, pos position) {
	if th, ok := t.(helperT); ok {
		th.Helper()
	}
	for i := 0; i < pos.structType.NumField(); i++ {
		if _, ok := cfg.ignored[pos.fieldName(i)]; ok {
			continue
		}
		modified := cfg.origFn()
		field := pos.getValue(modified).Field(i)

		switch f := reflect.Indirect(field); f.Kind() {
		case reflect.Struct:
			// TODO: limit depth

			nextPos := position{
				path:       pos.fieldName(i) + ".",
				structType: field.Type(),
				getValue: func(v reflect.Value) reflect.Value {
					return pos.getValue(v).Field(i)
				},
			}
			traverseStruct(t, cfg, nextPos)
		default:
			fillValue(cfg.rand, field)
			if !cfg.op(modified) {
				t.Fatalf("not complete: field %v is not included", pos.fieldName(i))
			}
		}
	}
}

func fillValue(rand *rand.Rand, target reflect.Value) {
	v := reflect.Indirect(target)

	if !v.CanSet() || v.Kind() == reflect.Invalid {
		panic(fmt.Sprintf("%v (%v) is not settable", v, v.Type()))
	}

	switch v.Kind() {
	case reflect.Bool:
		v.SetBool(!v.Bool())
	case reflect.String:
		v.SetString(randString(rand))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		for orig := v.Int(); orig == v.Int(); {
			v.SetInt(rand.Int63())
		}
	case reflect.Float32, reflect.Float64:
		for orig := v.Float(); orig == v.Float(); {
			v.SetFloat(rand.Float64())
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		for orig := v.Uint(); orig == v.Uint(); {
			v.SetUint(rand.Uint64())
		}
	case reflect.Complex64, reflect.Complex128:
		for orig := v.Complex(); orig == v.Complex(); {
			v.SetComplex(complex(rand.Float64(), rand.Float64()))
		}
	case reflect.Slice, reflect.Array:
		// TODO:
		panic("TODO: support slice and array")
	case reflect.Map:
		// TODO:
		panic("TODO: support map")
	case reflect.Struct:
		panic("structs should be filled by individual field")
	case reflect.Ptr, reflect.Interface:
		fallthrough
	case reflect.Chan, reflect.Func, reflect.UnsafePointer:
		panic(fmt.Sprintf("fill not implemented for kind %v", v.Kind()))
	}
}

var chars = []rune("你好欢迎abcdefghistuvwxyzABCDEFGHIJKLMNOμεταβλητόςпеременная")

func randString(rand *rand.Rand) string {
	length := rand.Intn(20) + 5
	var out strings.Builder
	for i := 0; i <= length; i++ {
		out.WriteRune(chars[rand.Intn(len(chars))])
	}
	return out.String()
}
