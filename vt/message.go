package vt

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"gotest.tools/v3/internal/source"
)

type msgResult struct {
	got        any
	want       any
	vtFuncName string
}

func Got(got any) string {
	// TODO colorize when in supported terminal
	const vtFuncName = "vt.Got"

	result := msgResult{got: got, vtFuncName: vtFuncName}

	// TODO: check if got is an error type from this package and return early

	callSource, err := getCallSource()
	if err != nil {
		// TODO: include tips about how to prevent this
		return fmt.Sprintf("%v, but %v: %v", result.basicMsg(), vtFuncName, err)
	}
	if len(callSource.CallExpr.Args) != 1 {
		return result.msgUnexpectedAstNode(callSource.CallExpr, "wrong number of args")
	}

	switch v := got.(type) {
	case nil:
		n, _ := source.FormatNode(callSource.CallExpr.Args[0])
		return fmt.Sprintf("%v is unable to produce a useful message, called with nil %v.",
			vtFuncName, n)
	case string:
		// diff from cmp.Diff
		// TODO:
	case error:
		// error comparison
		return result.handleSingleArgError(v, callSource)
	}
	// otherwise try for comparison to constant
	// TODO:

	return "TODO"
}

func GotWant(got any, want any) string {
	// TODO colorize when in supported terminal
	const vtFuncName = "vt.GotWant"

	result := msgResult{got: got}
	callSource, err := getCallSource()
	if err != nil {
		// TODO: include tips about how to prevent this
		return fmt.Sprintf("%v, but %v: %v", result.basicMsg(), vtFuncName, err)
	}
	if len(callSource.CallExpr.Args) != 2 {
		return result.msgUnexpectedAstNode(callSource.CallExpr, "wrong number of args")
	}

	// TODO: lookup comparison and switch on token

	// TODO:
	return "TODO"
}

func (r msgResult) basicMsg() string {
	var buf strings.Builder
	fmt.Fprintf(&buf, "got=%v", r.got)
	if r.want != nil {
		fmt.Fprintf(&buf, ", want=%v", r.want)
	}
	return buf.String()
}

func (r msgResult) handleSingleArgError(err error, callSource messageCallSource) string {
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
		return msg + fmt.Sprintf(" (%T), wanted %v", err, n)
	}

	return r.msgUnexpectedAstNode(ident, "unknown error comparison for variable")
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

func (r msgResult) msgErrorFromExpr(err error, expr ast.Expr, want ast.Expr) string {
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return r.msgUnexpectedAstNode(expr, "expected a function call for the variable declaration")
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

func (r msgResult) msgUnexpectedAstNode(node ast.Node, reason string) string {
	// TODO: include details about args, and request for a bug report
	n, _ := source.FormatNode(node)
	return fmt.Sprintf("%v: %v, got %T:\n%v", r.vtFuncName, reason, node, n)
}
