package property

import (
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"time"
)

// CompleteOptions are the settings used by Complete.
type CompleteOptions[T any] struct {
	// Operation should return true if the operation was successful, and false
	// otherwise. If Operation returns false, the test will be marked as failed
	// and the failure message will indicate which field of T was not used in
	// the operation.
	//
	// Operation should call the function being tested.
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

	// New is a function that returns a pointer to a struct type used by Operation.
	// If New is nil, the zero value of T will be used.
	//
	// New must return a pointer to a different instance each time it is called,
	// (it must be return a different pointer on every call).
	// New must always return a struct with the same values, otherwise the
	// assertion will fail. Generally New will return a struct literal populated
	// with hardcoded values.
	//
	// If the operation being tested is Empty (IsZero, IsEmpty, etc.) then New
	// should return the zero value of the struct. For other operations New
	// can set the fields to any value.
	//
	// The value returned by New will be used in two ways:
	//
	//   * the first value returned by New will be used as the first argument
	//     in every call to Operation.
	//   * the value returned by other calls to New will be modified and used as
	//     the second argument to Operation.
	//
	New func() T

	// Seed is the value used to initialize the random source. Defaults to
	// time.Time.UnixNano of the current time if unset. If a failure is only
	// reproducing with a specific seed, you can set this value to reproduce
	// the failure.
	Seed int64
}

// Complete tests that opt.Operation uses all the fields of struct T. See
// CompleteOptions for details about how to use Complete.
//
// Common operations that can be tested using Complete include:
//
//   - equal
//   - empty or isZero
//   - round tripping between two transformations
//   - building a hash from struct fields or map key for a struct
func Complete[T any](t TestingT, opt CompleteOptions[T]) {
	t.Helper()
	if opt.Seed == 0 {
		opt.Seed = time.Now().UnixNano()
	}
	t.Log("using random seed ", opt.Seed)

	newT := func() T {
		if opt.New == nil {
			return *new(T)
		}
		return opt.New()
	}
	orig := newT()
	cfg := config[T]{
		rand: rand.New(rand.NewSource(opt.Seed)),
		newT: newT,
		op: func(modified T) bool {
			return opt.Operation(orig, modified)
		},
		ignored: make(map[string]struct{}, len(opt.IgnoreFields)),
	}
	for _, k := range opt.IgnoreFields {
		cfg.ignored[k] = struct{}{}
	}
	pos := position{
		structType:     reflect.TypeOf(orig),
		getStructValue: func(v reflect.Value) reflect.Value { return v },
	}
	traverseStruct(t, cfg, pos)
}

// TestingT is the subset of testing.T used by functions in this package.
type TestingT interface {
	Log(args ...interface{})
	Fatalf(format string, args ...interface{})
	Helper()
}

// config is the internal version of CompleteOptions that is used by traverseStruct
type config[T any] struct {
	newT    func() T
	op      func(modified T) bool
	ignored map[string]struct{}
	rand    *rand.Rand
}

// position identifies the position in struct traversal.
type position struct {
	// path is the string representation of the position, used to compare to
	// keys in config.ignored, and as part of the failure message.
	path string
	// structType is the reflect.Type of struct at this position. The type is
	// used to lookup fields of the struct.
	structType reflect.Type
	// getStructValue is a function that receives a fresh copy of config.T from
	// config.newT, that is about to be modified. It returns the reflect.Value
	// for the struct at this position, which will be used to receive a random
	// value for the fields of the struct.
	getStructValue func(modified reflect.Value) reflect.Value
}

func traverseStruct[T any](t TestingT, cfg config[T], pos position) {
	t.Helper()
	for i := 0; i < pos.structType.NumField(); i++ {
		fieldType := pos.structType.Field(i)
		fieldPath := pos.path + fieldType.Name
		if _, ok := cfg.ignored[fieldPath]; ok {
			continue
		}
		modified := cfg.newT()
		field := pos.getStructValue(reflect.ValueOf(&modified).Elem()).Field(i)

		switch f := reflect.Indirect(field); f.Kind() {
		case reflect.Struct:
			// TODO: limit traversal depth to prevent infinite recursion

			nextPos := position{
				path:       fieldPath + ".",
				structType: field.Type(),
				getStructValue: func(v reflect.Value) reflect.Value {
					return pos.getStructValue(v).Field(i)
				},
			}
			traverseStruct(t, cfg, nextPos)
		default:
			fillValue(cfg.rand, field)
			if !cfg.op(modified) {
				t.Fatalf("not complete: field %v is not included", fieldPath)
			}
		}
	}
}

func fillValue(rand *rand.Rand, v reflect.Value) { //nolint:maintidx
	if v.Kind() == reflect.Pointer {
		v.Set(reflect.New(v.Type().Elem()))
		v = v.Elem()
	}

	if !v.CanSet() || v.Kind() == reflect.Invalid {
		panic(fmt.Sprintf("%v is not settable", v))
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
	case reflect.Slice:
		if v.Cap() == 0 || v.Len() == 0 {
			v.Set(reflect.MakeSlice(v.Type(), 1, 1))
		}
		fillValue(rand, v.Index(0))
	case reflect.Array:
		if v.Cap() > 0 {
			fillValue(rand, v.Index(0))
		}
	case reflect.Map:
		if v.Len() == 0 {
			v.Set(reflect.MakeMapWithSize(v.Type(), 1))
		}
		keyV := reflect.New(v.Type().Key()).Elem()
		fillValue(rand, keyV)
		valueV := reflect.New(v.Type().Elem()).Elem()
		fillValue(rand, valueV)
		v.SetMapIndex(keyV, valueV)
	case reflect.Struct:
		// TODO:
		panic("TODO: support struct in slice/map/array")
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
