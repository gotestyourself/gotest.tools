package assert

import (
	"bytes"
	"crypto/sha256"
	"strconv"
	"testing"
)

type exampleComplete struct {
	Field string
	Flag  bool
	Count int32
}

func (s exampleComplete) Equal(o exampleComplete) bool {
	return s.Field == o.Field && s.Flag == o.Flag && s.Count == o.Count
}

func (s exampleComplete) Empty() bool {
	return s.Count == 0 && !s.Flag && s.Field == ""
}

func (s exampleComplete) Key() []byte {
	h := sha256.New()
	h.Write([]byte(s.Field))
	h.Write([]byte(strconv.Itoa(int(s.Count))))
	h.Write([]byte(strconv.FormatBool(s.Flag)))
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
		in := Input[exampleComplete]{
			Original: func() *exampleComplete {
				return &exampleComplete{
					Field: "field-one",
					Flag:  true,
					Count: 3,
				}
			},
			Operation: func(x, y exampleComplete) bool {
				return !x.Equal(y)
			},
		}
		for i := 0; i < 200; i++ {
			Complete(t, in)
		}
	})
	t.Run("incomplete", func(t *testing.T) {
		in := Input[exampleIncomplete]{
			Original: func() *exampleIncomplete {
				return &exampleIncomplete{
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
		expectFailNowed(t, fakeT, "not complete: field Count is not included")
	})
	// TODO: test pointer fields
}

func TestComplete_WithKey(t *testing.T) {
	t.Run("complete", func(t *testing.T) {
		in := Input[exampleComplete]{
			Original: func() *exampleComplete {
				return &exampleComplete{
					Field: "field-one",
					Flag:  true,
					Count: 3,
				}
			},
			Operation: func(x, y exampleComplete) bool {
				return !bytes.Equal(x.Key(), y.Key())
			},
		}
		for i := 0; i < 200; i++ {
			Complete(t, in)
		}
	})
	t.Run("incomplete", func(t *testing.T) {
		in := Input[exampleIncomplete]{
			Original: func() *exampleIncomplete {
				return &exampleIncomplete{
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
		expectFailNowed(t, fakeT, "not complete: field Field is not included")
	})
}

func TestComplete_WithEmpty(t *testing.T) {
	t.Run("complete", func(t *testing.T) {
		in := Input[exampleComplete]{
			Original: func() *exampleComplete {
				return &exampleComplete{}
			},
			Operation: func(_, x exampleComplete) bool {
				return !x.Empty()
			},
		}
		for i := 0; i < 200; i++ {
			Complete(t, in)
		}
	})
	t.Run("incomplete", func(t *testing.T) {
		in := Input[exampleIncomplete]{
			Original: func() *exampleIncomplete {
				return &exampleIncomplete{}
			},
			Operation: func(_, x exampleIncomplete) bool {
				return !x.Empty()
			},
		}
		fakeT := &fakeTestingT{}
		Complete(fakeT, in)
		expectFailNowed(t, fakeT, "not complete: field Flag is not included")
	})
}

type exampleNested struct {
	Sub exampleIncomplete
	Top int8
}

func (s exampleNested) Equal(o exampleNested) bool {
	return s.Top == o.Top && s.Sub.Equal(o.Sub)
}

func TestComplete_Nested(t *testing.T) {
	t.Run("incomplete", func(t *testing.T) {
		in := Input[exampleNested]{
			Original: func() *exampleNested {
				return &exampleNested{
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
		expectFailNowed(t, fakeT, "not complete: field Sub.Count is not included")
	})
	// TODO: more test cases for nested (field as pointer to struct)
	// TODO: test cases for embedded
}
