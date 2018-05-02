package source_test

// using a separate package for test to avoid circular imports with the assert
// package

import (
	"testing"

	"gotest.tools/assert"
	"gotest.tools/internal/source"
)

func TestGetConditionSingleLine(t *testing.T) {
	msg, err := shim("not", "this", "this text")
	assert.NilError(t, err)
	assert.Equal(t, `"this text"`, msg)
}

func TestGetConditionMultiLine(t *testing.T) {
	msg, err := shim(
		"first",
		"second",
		"this text",
	)
	assert.NilError(t, err)
	assert.Equal(t, `"this text"`, msg)
}

func TestGetConditionIfStatement(t *testing.T) {
	if msg, err := shim(
		"first",
		"second",
		"this text",
	); true {
		assert.NilError(t, err)
		assert.Equal(t, `"this text"`, msg)
	}
}

func shim(_, _, _ string) (string, error) {
	return source.FormattedCallExprArg(1, 2)
}
