// Package source provides utilities for handling source-code.
package source // import "gotest.tools/v3/internal/source"

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"runtime"
)

// FormattedCallExprArg returns the argument from an ast.CallExpr at the
// index in the call stack. The argument is formatted using FormatNode.
func FormattedCallExprArg(stackIndex int, argPos int) (string, error) {
	args, err := CallExprArgs(stackIndex + 1)
	if err != nil {
		return "", err
	}
	if argPos >= len(args) {
		return "", errors.New("failed to find expression")
	}
	return FormatNode(args[argPos])
}

// CallExprArgs returns the ast.Expr slice for the args of an ast.CallExpr at
// the index in the call stack.
func CallExprArgs(stackIndex int) ([]ast.Expr, error) {
	_, filename, line, ok := runtime.Caller(stackIndex + 1)
	if !ok {
		return nil, errors.New("failed to get call stack")
	}
	debug("call stack position: %s:%d", filename, line)

	fs, err := ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to parse source file %s: %w", filename, err)
	}

	expr, err := GetCallExprArgs(fs, line)
	if err != nil {
		return nil, fmt.Errorf("call from %s:%d: %w", filename, line, err)
	}
	return expr, nil
}

func getNodeAtLine(fileset *token.FileSet, astFile ast.Node, lineNum int) (ast.Node, error) {
	if node := scanToLine(fileset, astFile, lineNum); node != nil {
		return node, nil
	}
	if node := scanToDeferLine(fileset, astFile, lineNum); node != nil {
		node, err := guessDefer(node)
		if err != nil || node != nil {
			return node, err
		}
	}
	return nil, fmt.Errorf("failed to find an expression")
}

func scanToLine(fileset *token.FileSet, node ast.Node, lineNum int) ast.Node {
	var matchedNode ast.Node
	ast.Inspect(node, func(node ast.Node) bool {
		switch {
		case node == nil || matchedNode != nil:
			return false
		case fileset.Position(node.Pos()).Line == lineNum:
			matchedNode = node
			return false
		}
		return true
	})
	return matchedNode
}

func GetCallExprArgs(src FileSource, line int) ([]ast.Expr, error) {
	node, err := getNodeAtLine(src.FileSet, src.AST, line)
	if err != nil {
		return nil, err
	}

	debug("found node: %s", debugFormatNode{node})

	// use Walk instead of Inspect because recursion is required
	visitor := &callExprVisitor{}
	ast.Walk(visitor, node)
	if visitor.expr == nil {
		return nil, errors.New("failed to find call expression")
	}
	debug("callExpr: %s", debugFormatNode{visitor.expr})
	return visitor.expr.Args, nil
}

type callExprVisitor struct {
	expr *ast.CallExpr
}

func (v *callExprVisitor) Visit(node ast.Node) ast.Visitor {
	if v.expr != nil || node == nil {
		return nil
	}
	debug("visit: %s", debugFormatNode{node})

	switch typed := node.(type) {
	case *ast.CallExpr:
		v.expr = typed
		return nil
	case *ast.DeferStmt:
		ast.Walk(v, typed.Call.Fun)
		return nil
	}
	return v
}

type FileSource struct {
	FileSet *token.FileSet
	AST     *ast.File
}

// TODO: cache with sync.DoOnce
func ReadFile(filename string) (FileSource, error) {
	fileset := token.NewFileSet()
	astFile, err := parser.ParseFile(fileset, filename, nil, parser.AllErrors|parser.ParseComments)
	if err != nil {
		return FileSource{}, fmt.Errorf("failed to read source file %s: %w", filename, err)
	}
	return FileSource{FileSet: fileset, AST: astFile}, nil
}

// FormatNode using go/format.Node and return the result as a string
func FormatNode(node ast.Node) (string, error) {
	buf := new(bytes.Buffer)
	err := format.Node(buf, token.NewFileSet(), node)
	return buf.String(), err
}

func debug(format string, args ...interface{}) {
	if os.Getenv("TEST_DEBUG") != "" {
		fmt.Fprintf(os.Stderr, "DEBUG: "+format+"\n", args...)
	}
}

type debugFormatNode struct {
	ast.Node
}

func (n debugFormatNode) String() string {
	if n.Node == nil {
		return "none"
	}
	out, err := FormatNode(n.Node)
	if err != nil {
		return fmt.Sprintf("failed to format %s: %s", n.Node, err)
	}
	return fmt.Sprintf("(%T) %s", n.Node, out)
}
