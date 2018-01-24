package source

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"runtime"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

const baseStackIndex = 1

// FormattedCallExprArg returns the argument from an ast.CallExpr at the
// index in the call stack. The argument is formatted using FormatNode.
func FormattedCallExprArg(stackIndex int, argPos int) (string, error) {
	args, err := CallExprArgs(stackIndex + 1)
	if err != nil {
		return "", err
	}
	return FormatNode(args[argPos])
}

func getNodeAtLine(filename string, lineNum int) (ast.Node, error) {
	fileset := token.NewFileSet()
	astFile, err := parser.ParseFile(fileset, filename, nil, parser.AllErrors)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse source file: %s", filename)
	}

	node := scanToLine(fileset, astFile, lineNum)
	if node == nil {
		return nil, errors.Errorf(
			"failed to find an expression on line %d in %s", lineNum, filename)
	}
	return node, nil
}

func scanToLine(fileset *token.FileSet, node ast.Node, lineNum int) ast.Node {
	v := &scanToLineVisitor{lineNum: lineNum, fileset: fileset}
	ast.Walk(v, node)
	return v.matchedNode
}

type scanToLineVisitor struct {
	lineNum     int
	matchedNode ast.Node
	fileset     *token.FileSet
}

func (v *scanToLineVisitor) Visit(node ast.Node) ast.Visitor {
	if node == nil || v.matchedNode != nil {
		return nil
	}
	if v.nodePosition(node).Line == v.lineNum {
		v.matchedNode = node
		return nil
	}
	return v
}

// In golang 1.9 the line number changed from being the line where the statement
// ended to the line where the statement began.
func (v *scanToLineVisitor) nodePosition(node ast.Node) token.Position {
	if isGOVersionBefore19() {
		return v.fileset.Position(node.End())
	}
	return v.fileset.Position(node.Pos())
}

func isGOVersionBefore19() bool {
	version := runtime.Version()
	// not a release version
	if !strings.HasPrefix(version, "go") {
		return false
	}
	version = strings.TrimPrefix(version, "go")
	parts := strings.Split(version, ".")
	if len(parts) < 2 {
		return false
	}
	minor, err := strconv.ParseInt(parts[1], 10, 32)
	return err == nil && parts[0] == "1" && minor < 9
}

func getCallExprArgs(node ast.Node) ([]ast.Expr, error) {
	visitor := &callExprVisitor{}
	ast.Walk(visitor, node)
	if visitor.expr == nil {
		return nil, errors.New("failed to find call expression")
	}
	return visitor.expr.Args, nil
}

type callExprVisitor struct {
	expr *ast.CallExpr
}

func (v *callExprVisitor) Visit(node ast.Node) ast.Visitor {
	if v.expr != nil || node == nil {
		return nil
	}

	switch typed := node.(type) {
	case *ast.CallExpr:
		switch typed.Fun.(type) {
		case *ast.Ident:
			v.expr = typed
			return nil
		}
	}
	return v
}

// FormatNode using go/format.Node and return the result as a string
func FormatNode(node ast.Node) (string, error) {
	buf := new(bytes.Buffer)
	err := format.Node(buf, token.NewFileSet(), node)
	return buf.String(), err
}

// CallExprArgs returns the ast.Expr slice for the args of an ast.CallExpr at
// the index in the call stack.
func CallExprArgs(stackIndex int) ([]ast.Expr, error) {
	_, filename, lineNum, ok := runtime.Caller(baseStackIndex + stackIndex)
	if !ok {
		return nil, errors.New("failed to get call stack")
	}

	node, err := getNodeAtLine(filename, lineNum)
	if err != nil {
		return nil, err
	}

	return getCallExprArgs(node)
}
