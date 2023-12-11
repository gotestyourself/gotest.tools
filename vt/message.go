package vt

import (
	"fmt"
	"go/ast"
	"runtime"

	"gotest.tools/v3/internal/source"
)

const vtMessageName = "vt.Message"

func Message(args ...any) string {
	nArgs := len(args)
	switch nArgs {
	case 0:
		return vtMessageName + "is unable to produce a useful message, called with no arguments."
	case 1:
	case 2:
		// TODO:
		return "TODO"
	default:
		return fmt.Sprintf("too many arguments in call to %v: %d", vtMessageName, nArgs)
	}

	// TODO: check custom error type for errors from this package

	_, filename, line, ok := runtime.Caller(1)
	if !ok {
		panic("failed to get call stack")
	}

	src, err := source.ReadFile(filename)
	if err != nil {
		// TODO: include details about args, and tips about how to prevent this
		return fmt.Sprintf("failed to lookup source: %v", err)
	}

	astArgs, err := source.GetCallExprArgs(src, line)
	if err != nil {
		// TODO: include details about args, and request for a bug report
		return fmt.Sprintf("failed to lookup call expression: %v", err)
	}

	switch v := args[0].(type) {
	case string:
		// diff from cmp.Diff
		// TODO:
	case error:
		// error from NilError
		return handleSingleArgError(v, astArgs[0], src)

	default:
		// TODO:
		_ = v
	}

	return "TODO"
}

func handleSingleArgError(err error, arg ast.Expr, src source.FileSource) string {
	ident, ok := arg.(*ast.Ident)
	if !ok {
		// TODO: include details about args, and request for a bug report
		n, _ := source.FormatNode(arg)
		return fmt.Sprintf("%v: expected a variable as the argument, got %T:\n%v", vtMessageName, arg, n)
	}

	switch v := ident.Obj.Decl.(type) {
	case *ast.AssignStmt:
		// TODO: handle multiple assignment
		call, ok := v.Rhs[0].(*ast.CallExpr)
		if !ok {
			return fmt.Sprintf("%v: expected a function call for the variable declaration, got %v",
				vtMessageName, v.Rhs[0])
		}

		// TODO: handle not an ident for the func
		n, _ := source.FormatNode(call)
		return fmt.Sprintf("%v returned an error: %v", n, err)

	case *ast.ValueSpec:
		// TODO: handle multiple declaration
	}
	return fmt.Sprintf(vtMessageName+"expected an assignment or a variable declaration for %v", ident.Name)
}
