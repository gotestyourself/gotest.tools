package vt

import (
	"fmt"
	"go/ast"
	"go/token"

	"gotest.tools/v3/internal/source"
)

func handleSingleArgString(v string, r msgResult, callSource messageCallSource) string {
	arg := callSource.CallExpr.Args[0]
	ident, ok := arg.(*ast.Ident)
	if !ok {
		return r.msgUnexpectedAstNode(arg, "expected a variable as the argument")
	}

	cond := callSource.IfStmt.Cond
	if condIsDiffIsNotEmpty(ident, cond) {
		cmpDiffCallExpr := exprFromObjDecl(ident)
		return r.msgStringFromExpr(v, cmpDiffCallExpr)
	}

	// TODO: what are other cases for Got(v string) ?
	return "TODO"
}

func condIsDiffIsNotEmpty(gotArg *ast.Ident, cond ast.Expr) bool {
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
	if xIdent.Name != gotArg.Name {
		return false
	}
	lit, ok := bExpr.Y.(*ast.BasicLit)
	if !ok {
		return false
	}
	return lit.Kind == token.STRING && lit.Value == `""`
}

func (r msgResult) msgStringFromExpr(diff string, cmpDiffCallExpr ast.Expr) string {
	ce, ok := cmpDiffCallExpr.(*ast.CallExpr)
	if !ok {
		return r.msgUnexpectedAstNode(cmpDiffCallExpr, "expected a function call for the variable declaration")
	}
	// TODO: check the call is cmp.Diff with the type system
	// TODO: check
	// TODO: handle want in first arg
	if len(ce.Args) < 2 {
		return r.msgUnexpectedAstNode(cmpDiffCallExpr, "expected a cmp.Diff function call with 2 or more args")
	}

	gotArg, ok := ce.Args[0].(*ast.Ident)
	if !ok {
		return r.msgUnexpectedAstNode(cmpDiffCallExpr, "expected a variable as arg to cmp.Diff")
	}

	callExpr := exprFromObjDecl(gotArg)
	call, ok := callExpr.(*ast.CallExpr)
	if !ok {
		return r.msgUnexpectedAstNode(callExpr, "expected a function call for the variable declaration")
	}

	// TODO: handle not an ident for the func
	// TODO: remove args if longer than x.
	n, _ := source.FormatNode(call)
	msg := fmt.Sprintf("%v returned a different result (-got +want):\n%v", n, diff)
	return msg
}
