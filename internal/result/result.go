// Package result provides helpers for handling comparison results.
package result

import (
	"fmt"
	"go/ast"
)

type withComparisonArgs interface {
	FailureMessage(args []ast.Expr) string
}

type basic interface {
	FailureMessage() string
}

// FailureMessage returns the failure message from result. If result requires
// comparison arguments, args are filtered to only include expressions that
// are useful when printed.
func FailureMessage(result interface{}, args []ast.Expr) string {
	switch typed := result.(type) {
	case withComparisonArgs:
		return typed.FailureMessage(filterPrintableExpr(args))
	case basic:
		return typed.FailureMessage()
	default:
		return fmt.Sprintf("comparison returned invalid Result type: %T", result)
	}
}

// UsesArgs reports whether result requires comparison arguments to
// produce its failure message.
func UsesArgs(result interface{}) bool {
	_, ok := result.(withComparisonArgs)
	return ok
}

// filterPrintableExpr filters the ast.Expr slice to only include Expr that are
// easy to read when printed and contain relevant information to an assertion.
//
// Ident and SelectorExpr are included because they print nicely and the variable
// names may provide additional context to their values.
// BasicLit and CompositeLit are excluded because their source is equivalent to
// their value, which is already available.
// Other types are ignored for now, but could be added if they are relevant.
func filterPrintableExpr(args []ast.Expr) []ast.Expr {
	res := make([]ast.Expr, len(args))
	for i, arg := range args {
		if isShortPrintableExpr(arg) {
			res[i] = arg
			continue
		}

		if starExpr, ok := arg.(*ast.StarExpr); ok {
			res[i] = starExpr.X
			continue
		}
	}
	return res
}

func isShortPrintableExpr(expr ast.Expr) bool {
	switch expr.(type) {
	case *ast.Ident, *ast.SelectorExpr, *ast.IndexExpr, *ast.SliceExpr:
		return true
	case *ast.BinaryExpr, *ast.UnaryExpr:
		return true
	default:
		// CallExpr, ParenExpr, TypeAssertExpr, KeyValueExpr, StarExpr
		return false
	}
}
