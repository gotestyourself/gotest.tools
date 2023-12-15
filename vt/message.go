package vt

import (
	"fmt"
	"go/ast"
	"go/token"
	"runtime"

	"gotest.tools/v3/internal/source"
)

const vtMessageName = "vt.Message"

func Message(got any, optionalWant ...any) string {
	var want any
	var hasWants int
	var wantV string
	switch hasWants = len(optionalWant); hasWants {
	case 0:
	case 1:
		want = optionalWant[0]
		wantV = fmt.Sprintf("%v", want)
	default:
		wantV = fmt.Sprintf("%v", optionalWant)
	}

	// TODO: check if got is an error type from this package and return early

	_, filename, line, ok := runtime.Caller(1)
	if !ok {
		panic("failed to get call stack")
	}
	src, err := source.ReadFile(filename)
	if err != nil {
		// TODO: include tips about how to prevent this, auto-fix?
		return fmt.Sprintf("got=%v, want=%v but %v failed to lookup source: %v",
			got, wantV, vtMessageName, err)
	}

	callSource, err := getNodeAtLine(src, line)
	if err != nil {
		// TODO: include details about args, and request for a bug report
		return fmt.Sprintf("failed to lookup call expression: %v", err)
	}

	if len(optionalWant) > 1 {
		// TODO: print warning about too many args instead of exit
		// TODO: auto-fix
		return "too many optionalWant for " + vtMessageName
	}
	if len(callSource.CallExpr.Args) != 1+hasWants {
		return msgUnexpectedAstNode(callSource.CallExpr, "wrong number of args")
	}
	switch v := got.(type) {
	case nil:
		n, _ := source.FormatNode(callSource.CallExpr.Args[0])
		return fmt.Sprintf("%v is unable to produce a useful message, called with nil %v.",
			vtMessageName, n)
	case string:
		// diff from cmp.Diff
		// TODO:
	case error:
		// error comparison
		return handleSingleArgError(v, want, callSource)

	default:
		// TODO:
	}

	return "TODO"
}

func handleSingleArgError(err error, want any, callSource messageCallSource) string {
	arg := callSource.CallExpr.Args[0]
	ident, ok := arg.(*ast.Ident)
	if !ok {
		return msgUnexpectedAstNode(arg, "expected a variable as the argument")
	}

	cond := callSource.IfStmt.Cond
	if condIsErrNotNil(ident, cond) {
		callExpr := exprFromObjDecl(ident)
		return msgErrorFromExpr(err, callExpr, nil)
	}

	if wantExpr, ok := condIsNotErrorsIs(ident, cond); ok {
		callExpr := exprFromObjDecl(ident)
		return msgErrorFromExpr(err, callExpr, wantExpr)
	}

	if wantExpr, ok := condIsNotErrorsAs(ident, cond); ok {
		callExpr := exprFromObjDecl(ident)
		msg := msgErrorFromExpr(err, callExpr, nil)
		n, _ := source.FormatNode(wantExpr)
		return msg + fmt.Sprintf(" (%T), wanted %v", err, n)
	}

	return msgUnexpectedAstNode(ident, "unknown error comparison for variable")
}

func exprFromObjDecl(ident *ast.Ident) ast.Expr {
	switch v := ident.Obj.Decl.(type) {
	case *ast.AssignStmt:
		// TODO: handle multiple assignment
		return v.Rhs[0]

	case *ast.ValueSpec:
		// TODO: handle multiple declaration
		return v.Values[0]
	}
	// TODO: other cases?
	return nil
}

// !errors.Is(err, want)
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

// !errors.As(err, want)
func condIsNotErrorsAs(errArg *ast.Ident, cond ast.Expr) (ast.Expr, bool) {
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
	if se.Sel.Name != "As" {
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

	// unwrap if the expr is &typedErr
	want := ce.Args[1]
	if uExpr, ok := want.(*ast.UnaryExpr); ok {
		if uExpr.Op != token.AND {
			return nil, false
		}
		want = uExpr.X
	}

	wantIdent, ok := want.(*ast.Ident)
	if !ok {
		// TODO: what else could it be?
		return nil, false
	}

	// TODO: Use the type system here
	declExpr := exprFromObjDecl(wantIdent)
	// unwrap the declaration
	// TODO: extract func
	if uExpr, ok := declExpr.(*ast.UnaryExpr); ok {
		if uExpr.Op != token.AND {
			return nil, false
		}
		declExpr = uExpr.X
	}

	cl, ok := declExpr.(*ast.CompositeLit)
	if !ok {
		return nil, false
	}

	want, ok = cl.Type.(*ast.Ident)
	if !ok {
		return nil, false
	}
	return want, true
}

// err != nil
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
