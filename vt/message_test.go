package vt

import (
	"fmt"
	"testing"
)

func TestMessage(t *testing.T) {
	someFunc := func(...any) error {
		return fmt.Errorf("failed to do something")
	}

	t.Run("err: assignment from function", func(t *testing.T) {
		var got string
		if err := someFunc("a", 1, nil); err != nil {
			got = Message(err)
		}

		want := `someFunc("a", 1, nil) returned an error: failed to do something`
		if got != want {
			t.Fatalf("Message(err)\ngot:  %v\nwant: %v", got, want)
		}
	})
}
