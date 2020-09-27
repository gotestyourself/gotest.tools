package source_test

// using a separate package for test to avoid circular imports with the assert
// package

import (
	"fmt"
	"runtime"
	"strings"
	"testing"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/internal/source"
	"gotest.tools/v3/skip"
)

func TestFormattedCallExprArg_SingleLine(t *testing.T) {
	gotestToolsTestShim := &capture{argPos: 2}
	msg, err := gotestToolsTestShim.shim("not", "this", "this text")
	assert.NilError(t, err)
	assert.Equal(t, `"this text"`, msg)
}

func TestFormattedCallExprArg_MultiLine(t *testing.T) {
	gotestToolsTestShim := &capture{argPos: 2}
	msg, err := gotestToolsTestShim.shim(
		"first",
		"second",
		"this text",
	)
	assert.NilError(t, err)
	assert.Equal(t, `"this text"`, msg)
}

func TestFormattedCallExprArg_IfStatement(t *testing.T) {
	gotestToolsTestShim := &capture{argPos: 2}
	if msg, err := gotestToolsTestShim.shim(
		"first",
		"second",
		"this text",
	); true {
		assert.NilError(t, err)
		assert.Equal(t, `"this text"`, msg)
	}
}

func TestFormattedCallExprArg_InDefer(t *testing.T) {
	skip.If(t, isGoVersion18)
	gotestToolsTestShim := &capture{argPos: 1}
	func() {
		defer gotestToolsTestShim.shim("first", "second")
	}()

	assert.NilError(t, gotestToolsTestShim.err)
	assert.Equal(t, gotestToolsTestShim.value, `"second"`)
}

func isGoVersion18() bool {
	return strings.HasPrefix(runtime.Version(), "go1.8.")
}

type capture struct {
	argPos int
	value  string
	err    error
}

func (c *capture) shim(_ ...string) (string, error) {
	c.value, c.err = source.FormattedCallExprArg(1, c.argPos)
	return c.value, c.err
}

func TestFormattedCallExprArg_InAnonymousDefer(t *testing.T) {
	gotestToolsTestShim := &capture{argPos: 1}
	func() {
		fmt.Println()
		defer fmt.Println()
		defer func() { gotestToolsTestShim.shim("first", "second") }()
	}()

	assert.NilError(t, gotestToolsTestShim.err)
	assert.Equal(t, gotestToolsTestShim.value, `"second"`)
}

func TestFormattedCallExprArg_InDeferMultipleDefers(t *testing.T) {
	skip.If(t, isGoVersion18)
	gotestToolsTestShim := &capture{argPos: 1}
	func() {
		fmt.Println()
		defer fmt.Println()
		defer gotestToolsTestShim.shim("first", "second")
	}()

	assert.ErrorContains(t, gotestToolsTestShim.err, "ambiguous call expression")
}
