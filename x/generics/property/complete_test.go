package property

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"gotest.tools/v3/assert"
)

type exampleComplete struct {
	Field string
	Flag  bool
	Count int32
	Maybe *string
	Array [2]int
	Seq   []byte
	Assoc map[int]string
}

func (s exampleComplete) Equal(o exampleComplete) bool {
	if s.Array != o.Array {
		return false
	}
	if !bytes.Equal(s.Seq, o.Seq) {
		return false
	}
	if !reflect.DeepEqual(s.Assoc, o.Assoc) {
		return false
	}
	return s.Field == o.Field && s.Flag == o.Flag && s.Count == o.Count && s.Maybe == o.Maybe
}

func (s exampleComplete) Empty() bool {
	return s.Equal(exampleComplete{})
}

func (s exampleComplete) Key() []byte {
	h := sha256.New()
	h.Write([]byte(s.Field))
	h.Write([]byte(strconv.Itoa(int(s.Count))))
	h.Write([]byte(strconv.FormatBool(s.Flag)))
	if s.Maybe != nil {
		h.Write([]byte(*s.Maybe))
	}
	for _, item := range s.Array {
		h.Write([]byte(strconv.Itoa(item)))
	}
	h.Write(s.Seq)
	for k, v := range s.Assoc {
		h.Write([]byte(strconv.Itoa(k)))
		h.Write([]byte(v))
	}
	return h.Sum(nil)
}

type exampleIncomplete struct {
	Field string
	Flag  bool
	Count int32
}

func (s exampleIncomplete) Equal(o exampleIncomplete) bool {
	return s.Field == o.Field && s.Flag == o.Flag
}

func (s exampleIncomplete) Empty() bool {
	return s.Count == 0 && s.Field == ""
}

func (s exampleIncomplete) Key() []byte {
	h := sha256.New()
	h.Write([]byte(strconv.Itoa(int(s.Count))))
	h.Write([]byte(strconv.FormatBool(s.Flag)))
	return h.Sum(nil)
}

func TestComplete_WithEqual(t *testing.T) {
	t.Run("complete", func(t *testing.T) {
		in := CompleteOptions[exampleComplete]{
			New: func() exampleComplete {
				return exampleComplete{
					Field: "field-one",
					Flag:  true,
					Count: 3,
					Array: [2]int{1, 2},
				}
			},
			Operation: func(x, y exampleComplete) bool {
				return !x.Equal(y)
			},
			IgnoreFields: []string{"Assoc"},
		}
		for i := 0; i < 200; i++ {
			Complete(t, in)
		}
	})
	t.Run("incomplete", func(t *testing.T) {
		in := CompleteOptions[exampleIncomplete]{
			New: func() exampleIncomplete {
				return exampleIncomplete{
					Field: "field-one",
					Flag:  true,
					Count: 3,
				}
			},
			Operation: func(x, y exampleIncomplete) bool {
				return !x.Equal(y)
			},
		}
		fakeT := &fakeTestingT{}
		Complete(fakeT, in)
		expectCompleteFailure(t, fakeT, "not complete: field Count is not included")
	})
	t.Run("complete with ignore fields", func(t *testing.T) {
		in := CompleteOptions[exampleIncomplete]{
			New: func() exampleIncomplete {
				return exampleIncomplete{
					Field: "field-one",
					Flag:  true,
					Count: 3,
				}
			},
			Operation: func(x, y exampleIncomplete) bool {
				return !x.Equal(y)
			},
			IgnoreFields: []string{"Count"},
		}
		for i := 0; i < 200; i++ {
			Complete(t, in)
		}
	})
	t.Run("complete with default new", func(t *testing.T) {
		in := CompleteOptions[exampleComplete]{
			Operation: func(x, y exampleComplete) bool {
				return !x.Equal(y)
			},
			IgnoreFields: []string{"Count", "Assoc"},
		}
		for i := 0; i < 200; i++ {
			Complete(t, in)
		}
	})
	t.Run("incomplete with default new", func(t *testing.T) {
		in := CompleteOptions[exampleIncomplete]{
			Operation: func(x, y exampleIncomplete) bool {
				return !x.Equal(y)
			},
		}
		fakeT := &fakeTestingT{}
		Complete(fakeT, in)
		expectCompleteFailure(t, fakeT, "not complete: field Count is not included")
	})
}

func expectCompleteFailure(t *testing.T, fakeT *fakeTestingT, expected string) {
	t.Helper()
	assert.Assert(t, fakeT.failed, "should have failed")
	if len(fakeT.msgs) < 2 {
		t.Fatalf("expected at least 2 log messages: %v", fakeT.msgs)
	}
	assert.Equal(t, fakeT.msgs[1], expected, "wrong failure message")
}

func TestComplete_WithKey(t *testing.T) {
	t.Run("complete", func(t *testing.T) {
		in := CompleteOptions[exampleComplete]{
			New: func() exampleComplete {
				return exampleComplete{
					Field: "field-one",
					Flag:  true,
					Count: 3,
				}
			},
			Operation: func(x, y exampleComplete) bool {
				return !bytes.Equal(x.Key(), y.Key())
			},
			IgnoreFields: []string{"Assoc"},
		}
		for i := 0; i < 200; i++ {
			Complete(t, in)
		}
	})
	t.Run("incomplete", func(t *testing.T) {
		in := CompleteOptions[exampleIncomplete]{
			New: func() exampleIncomplete {
				return exampleIncomplete{
					Field: "field-one",
					Flag:  true,
					Count: 3,
				}
			},
			Operation: func(x, y exampleIncomplete) bool {
				return !bytes.Equal(x.Key(), y.Key())
			},
		}
		fakeT := &fakeTestingT{}
		Complete(fakeT, in)
		expectCompleteFailure(t, fakeT, "not complete: field Field is not included")
	})
}

func TestComplete_WithEmpty(t *testing.T) {
	t.Run("complete", func(t *testing.T) {
		in := CompleteOptions[exampleComplete]{
			New: func() exampleComplete {
				return exampleComplete{}
			},
			Operation: func(_, x exampleComplete) bool {
				return !x.Empty()
			},
			IgnoreFields: []string{"Assoc"},
		}
		for i := 0; i < 200; i++ {
			Complete(t, in)
		}
	})
	t.Run("incomplete", func(t *testing.T) {
		in := CompleteOptions[exampleIncomplete]{
			New: func() exampleIncomplete {
				return exampleIncomplete{}
			},
			Operation: func(_, x exampleIncomplete) bool {
				return !x.Empty()
			},
		}
		fakeT := &fakeTestingT{}
		Complete(fakeT, in)
		expectCompleteFailure(t, fakeT, "not complete: field Flag is not included")
	})
}

type exampleNested struct {
	Sub exampleIncomplete
	Top int8
	//Assoc map[exampleKey]exampleValue
	Seq []exampleValue

	exampleValue
}

func (s exampleNested) Equal(o exampleNested) bool {
	if s.Top != o.Top || !s.Sub.Equal(o.Sub) {
		return false
	}
	for i := range s.Seq {
		if len(o.Seq) <= i || s.Seq[i] != o.Seq[i] {
			return false
		}
	}
	if len(s.Seq) != len(o.Seq) {
		return false
	}
	return s.exampleValue.One == o.exampleValue.One
}

type exampleValue struct {
	One int
	Ok  bool
}

func TestComplete_Nested(t *testing.T) {
	t.Run("incomplete", func(t *testing.T) {
		in := CompleteOptions[exampleNested]{
			New: func() exampleNested {
				return exampleNested{
					Sub: exampleIncomplete{
						Field: "what",
						Flag:  true,
						Count: 23,
					},
					Top: 12,
				}
			},
			Operation: func(x, y exampleNested) bool {
				return !x.Equal(y)
			},
		}
		fakeT := &fakeTestingT{}
		Complete(fakeT, in)
		expectCompleteFailure(t, fakeT, "not complete: field Sub.Count is not included")
	})
	t.Run("complete with ignore fields", func(t *testing.T) {
		in := CompleteOptions[exampleNested]{
			New: func() exampleNested {
				return exampleNested{
					Sub: exampleIncomplete{
						Field: "what",
						Flag:  true,
						Count: 23,
					},
					Top: 12,
				}
			},
			Operation: func(x, y exampleNested) bool {
				return !x.Equal(y)
			},
			IgnoreFields: []string{"Sub.Count", "exampleValue.Ok"},
		}
		for i := 0; i < 200; i++ {
			Complete(t, in)
		}
	})

	// TODO: more test cases for nested (field as pointer to struct)
}

type fakeTestingT struct {
	failed bool
	msgs   []string
}

func (f *fakeTestingT) Log(args ...interface{}) {
	f.msgs = append(f.msgs, fmt.Sprint(args...))
}

func (f *fakeTestingT) Helper() {}

func (f *fakeTestingT) Fatalf(format string, args ...interface{}) {
	f.failed = true
	f.msgs = append(f.msgs, fmt.Sprintf(format, args...))
}
