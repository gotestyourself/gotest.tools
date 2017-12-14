package source_test

// using a separate package for test to avoid circular imports with the assert
// package

import (
	"testing"

	"github.com/gotestyourself/gotestyourself/assert"
	"github.com/gotestyourself/gotestyourself/internal/source"
)

func TestGetConditionSingleLine(t *testing.T) {
	msg, err := shim("not", "this", "this text")
	assert.NoError(t, err)
	assert.Equal(t, `"this text"`, msg)
}

func TestGetConditionMultiLine(t *testing.T) {
	msg, err := shim(
		"first",
		"second",
		"this text",
	)
	assert.NoError(t, err)
	assert.Equal(t, `"this text"`, msg)
}

func TestGetConditionIfStatement(t *testing.T) {
	if msg, err := shim(
		"first",
		"second",
		"this text",
	); true {
		assert.NoError(t, err)
		assert.Equal(t, `"this text"`, msg)
	}
}

func shim(_, _, _ string) (string, error) {
	return source.GetCondition(1, 2)
}
