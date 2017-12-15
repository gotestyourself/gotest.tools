package source

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetConditionSingleLine(t *testing.T) {
	msg, err := shim("not", "this", "this text")
	require.NoError(t, err)
	assert.Equal(t, `"this text"`, msg)
}

func TestGetConditionMultiLine(t *testing.T) {
	msg, err := shim(
		"first",
		"second",
		"this text",
	)
	require.NoError(t, err)
	assert.Equal(t, `"this text"`, msg)
}

func TestGetConditionIfStatement(t *testing.T) {
	if msg, err := shim(
		"first",
		"second",
		"this text",
	); true {
		require.NoError(t, err)
		assert.Equal(t, `"this text"`, msg)
	}
}

func shim(_, _, _ string) (string, error) {
	return GetCondition(1, 2)
}
