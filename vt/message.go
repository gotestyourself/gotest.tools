package vt

import (
	"fmt"
	"go/ast"
	"go/token"
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

	callSource, err := getNodeAtLine(src, line)
	if err != nil {
		// TODO: include details about args, and request for a bug report
		return fmt.Sprintf("failed to lookup call expression: %v", err)
	}

	if len(callSource.CallExpr.Args) != len(args) {
		return msgUnexpectedAstNode(callSource.CallExpr, "wrong number of args")
	}

	switch v := args[0].(type) {
	case string:
		// diff from cmp.Diff
		// TODO:
	case error:
		// error from NilError
		return handleSingleArgError(v, callSource)

	default:
		// TODO:
		_ = v
	}

	return "TODO"
}

func handleSingleArgError(err error, callSource messageCallSource) string {
	arg := callSource.CallExpr.Args[0]
	ident, ok := arg.(*ast.Ident)
	if !ok {
		return msgUnexpectedAstNode(arg, "expected a variable as the argument")
	}

	if condIsErrNotNil(ident, callSource.IfStmt.Cond) {
		switch v := ident.Obj.Decl.(type) {
		case *ast.AssignStmt:
			// TODO: handle multiple assignment
			return msgErrorFromExpr(err, v.Rhs[0], nil)

		case *ast.ValueSpec:
			// TODO: handle multiple declaration
			return msgErrorFromExpr(err, v.Values[0], nil)

		}
	}

	if wantExpr, ok := condIsNotErrorsIs(ident, callSource.IfStmt.Cond); ok {
		switch v := ident.Obj.Decl.(type) {
		case *ast.AssignStmt:
			// TODO: handle multiple assignment
			return msgErrorFromExpr(err, v.Rhs[0], wantExpr)

		case *ast.ValueSpec:
			// TODO: handle multiple declaration
			return msgErrorFromExpr(err, v.Values[0], wantExpr)

		}
	}

	return msgUnexpectedAstNode(ident, "expected an assignment or a variable declaration")
}

func condIsNotErrorsIs(errArg *ast.Ident, cond ast.Expr) (ast.Expr, bool) {
	uExpr, ok := cond.(*ast.UnaryExpr)
	if !ok {
		return nil, false
	}
	if uExpr.Op != token.NOT {
		return nil, false
	}
	ce, ok := uExpr.X.(*ast.CallExpr)
	if !ok {
		return nil, false
	}

	se, ok := ce.Fun.(*ast.SelectorExpr)
	if !ok {
		return nil, false
	}

	x, ok := se.X.(*ast.Ident)
	if !ok {
		return nil, false
	}
	if x.Name != "errors" {
		return nil, false
	}
	if se.Sel.Name != "Is" {
		return nil, false
	}

	if len(ce.Args) != 2 {
		return nil, false
	}

	arg0, ok := ce.Args[0].(*ast.Ident)
	if !ok {
		return nil, false
	}
	if arg0.Name != errArg.Name {
		return nil, false
	}

	return ce.Args[1], true
}

func condIsErrNotNil(errArg *ast.Ident, cond ast.Expr) bool {
	bExpr, ok := cond.(*ast.BinaryExpr)
	if !ok {
		return false
	}
	if bExpr.Op != token.NEQ {
		return false
	}
	xIdent, ok := bExpr.X.(*ast.Ident)
	if !ok {
		return false
	}
	if xIdent.Name != errArg.Name {
		return false
	}
	yIdent, ok := bExpr.Y.(*ast.Ident)
	if !ok {
		return false
	}
	return yIdent.Name == "nil"
}

func msgErrorFromExpr(err error, expr ast.Expr, want ast.Expr) string {
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return msgUnexpectedAstNode(expr, "expected a function call for the variable declaration")
	}

	// TODO: handle not an ident for the func
	// TODO: remove args if longer than x.
	n, _ := source.FormatNode(call)
	msg := fmt.Sprintf("%v returned error: %v", n, err)
	if want != nil {
		n, _ = source.FormatNode(want)
		msg += fmt.Sprintf(", wanted %v", n)
	}
	return msg
}

func msgUnexpectedAstNode(node ast.Node, reason string) string {
	// TODO: include details about args, and request for a bug report
	n, _ := source.FormatNode(node)
	return fmt.Sprintf("%v: %v, got %T:\n%v", vtMessageName, reason, node, n)
}
