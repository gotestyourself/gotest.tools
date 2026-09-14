package assert

import (
	"errors"
	"go/ast"

	"gotest.tools/v3/assert/cmp"
	"gotest.tools/v3/internal/format"
	"gotest.tools/v3/internal/result"
	"gotest.tools/v3/internal/source"
)

// RunComparison and return Comparison.Success. If the comparison fails a messages
// will be printed using t.Log.
func RunComparison(
	t LogT,
	argSelector argSelector,
	f cmp.Comparison,
	msgAndArgs ...interface{},
) bool {
	if ht, ok := t.(helperT); ok {
		ht.Helper()
	}
	res := f()
	if res.Success() {
		return true
	}

	if source.IsUpdate() {
		if updater, ok := res.(updateExpected); ok {
			const stackIndex = 3 // Assert/Check, assert, RunComparison
			err := updater.UpdatedExpected(stackIndex)
			switch {
			case err == nil:
				return true
			case errors.Is(err, source.ErrNotFound):
				// do nothing, fallthrough to regular failure message
			default:
				t.Log("failed to update source", err)
				return false
			}
		}
	}

	var args []ast.Expr
	if result.UsesArgs(res) {
		const stackIndex = 3 // Assert/Check, assert, RunComparison
		var err error
		args, err = source.CallExprArgs(stackIndex)
		if err != nil {
			t.Log(err.Error())
		}
		args = argSelector(args)
	}

	message := result.FailureMessage(res, args)
	t.Log(format.WithCustomMessage(failureMessage+message, msgAndArgs...))
	return false
}

type updateExpected interface {
	UpdatedExpected(stackIndex int) error
}

type argSelector func([]ast.Expr) []ast.Expr

// ArgsAfterT selects args starting at position 1. Used when the caller has a
// testing.T as the first argument, and the args to select should follow it.
func ArgsAfterT(args []ast.Expr) []ast.Expr {
	if len(args) < 1 {
		return nil
	}
	return args[1:]
}

// ArgsFromComparisonCall selects args from the CallExpression at position 1.
// Used when the caller has a testing.T as the first argument, and the args to
// select are passed to the cmp.Comparison at position 1.
func ArgsFromComparisonCall(args []ast.Expr) []ast.Expr {
	if len(args) <= 1 {
		return nil
	}
	if callExpr, ok := args[1].(*ast.CallExpr); ok {
		return callExpr.Args
	}
	return nil
}

// ArgsAtZeroIndex selects args from the CallExpression at position 1.
// Used when the caller accepts a single cmp.Comparison argument.
func ArgsAtZeroIndex(args []ast.Expr) []ast.Expr {
	if len(args) == 0 {
		return nil
	}
	if callExpr, ok := args[0].(*ast.CallExpr); ok {
		return callExpr.Args
	}
	return nil
}
