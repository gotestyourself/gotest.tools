package vt

import (
	"fmt"
	"go/ast"
	"go/token"

	"gotest.tools/v3/internal/source"
)

func handleSingleArgError(err error, r msgResult, callSource messageCallSource) string {
	arg := callSource.CallExpr.Args[0]
	ident, ok := arg.(*ast.Ident)
	if !ok {
		return r.msgUnexpectedAstNode(arg, "expected a variable as the argument")
	}

	cond := callSource.IfStmt.Cond
	if condIsErrNotNil(ident, cond) {
		callExpr := exprFromObjDecl(ident)
		return r.msgErrorFromExpr(err, callExpr, nil)
	}

	if wantExpr, ok := condIsNotErrorsIs(ident, cond); ok {
		callExpr := exprFromObjDecl(ident)
		return r.msgErrorFromExpr(err, callExpr, wantExpr)
	}

	if wantExpr, ok := condIsNotErrorsAs(ident, cond); ok {
		callExpr := exprFromObjDecl(ident)
		msg := r.msgErrorFromExpr(err, callExpr, nil)
		n, _ := source.FormatNode(wantExpr)
		// TODO: this breaks with comments, find a better way to include the type of err
		return msg + fmt.Sprintf(" (%T), wanted %v", err, n)
	}

	return r.msgUnexpectedAstNode(ident, "unknown error comparison for variable")
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
