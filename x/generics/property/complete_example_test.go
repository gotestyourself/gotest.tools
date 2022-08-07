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
	return f.Kind == "" && f.Quantity == 0
}

func ExampleComplete() {
	var t *testing.T // for example only

	// This subtest will fail if someone adds a new field to FoodRequest
	// and forgets to change the Equal method to include the field.
	t.Run("Equal is complete", func(t *testing.T) {
		property.Complete(t, property.CompleteOptions[FoodRequest]{
			New: func() *FoodRequest {
				return &FoodRequest{Kind: "apple", Quantity: 3}
			},
			Operation: func(x, y FoodRequest) bool {
				return x.Equal(y)
			},
		})
	})

	// This subtest will fail if someone adds a new field to FoodRequest
	// and forgets to change the IsZero method to include the field.
	t.Run("IsZero is complete", func(t *testing.T) {
		property.Complete(t, property.CompleteOptions[FoodRequest]{
			New: func() *FoodRequest {
				return &FoodRequest{}
			},
			Operation: func(_, y FoodRequest) bool {
				return !y.IsZero()
			},
		})
	})
}
