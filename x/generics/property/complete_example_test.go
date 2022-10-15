package property_test

import (
	"testing"

	"gotest.tools/x/generics/property"
)

type FoodRequest struct {
	Kind     string
	Quantity int
}

func (f FoodRequest) Equal(o FoodRequest) bool {
	return f.Kind == o.Kind && f.Quantity == o.Quantity
}

func (f FoodRequest) IsZero() bool {
	return f.Equal(FoodRequest{})
}

func ExampleComplete() {
	var t *testing.T // for example only

	// This subtest will fail if someone adds a new field to FoodRequest
	// and forgets to change the Equal method to include the field.
	t.Run("Equal is complete", func(t *testing.T) {
		property.Complete(t, property.CompleteOptions[FoodRequest]{
			Operation: func(x, y FoodRequest) bool {
				// when any field is changed, the two values should not be equal
				return !x.Equal(y)
			},
		})
	})

	// This subtest will fail if someone adds a new field to FoodRequest
	// and forgets to change the IsZero method to include the field.
	t.Run("IsZero is complete", func(t *testing.T) {
		property.Complete(t, property.CompleteOptions[FoodRequest]{
			Operation: func(_, modified FoodRequest) bool {
				// when any field is changed, the modified value should no longer
				// be the zero value.
				return !modified.IsZero()
			},
		})
	})
}
