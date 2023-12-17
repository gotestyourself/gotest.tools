package vt

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"os"
	"runtime"

	"gotest.tools/v3/internal/source"
)

func getCallSource() (messageCallSource, error) {
	_, filename, line, ok := runtime.Caller(2)
	if !ok {
		panic("failed to get call stack")
	}
	src, err := source.ReadFile(filename)
	if err != nil {
		return messageCallSource{}, fmt.Errorf("failed to read Go source file: %w", err)
	}

	callSource, err := getNodeAtLine(src, line)
	if err != nil {
		return messageCallSource{}, fmt.Errorf("failed to lookup call expression: %w", err)
	}
	return callSource, nil
}

type messageCallSource struct {
	FileSet        *token.FileSet
	File           *ast.File
	CallExpr       *ast.CallExpr
	CallComments   []*ast.CommentGroup
	IfStmt         *ast.IfStmt
	IfStmtComments []*ast.CommentGroup
}

// TODO: defer support was removed
func getNodeAtLine(src source.FileSource, lineNum int) (messageCallSource, error) {
	result := scanToLine(src, lineNum)
	if result.CallExpr == nil || result.IfStmt == nil {
		return result, fmt.Errorf("failed to find an expression on line")
	}
	debug("found node: %s", debugFormatNode{result.IfStmt})
	return result, nil
}

func scanToLine(src source.FileSource, lineNum int) messageCallSource {
	fileset := src.FileSet

	cmap := ast.NewCommentMap(src.FileSet, src.AST, src.AST.Comments)

	var result messageCallSource
	result.File = src.AST
	result.FileSet = src.FileSet

	pre := func(current ast.Node) bool {
		if current == nil || result.CallExpr != nil {
			return false
		}

		if fileset.Position(current.Pos()).Line > lineNum {
			return false // past the relevant scope
		}

		if fileset.Position(current.End()).Line < lineNum {
			return true // before the relevant scope
		}

		if ifStmt, ok := current.(*ast.IfStmt); ok {
			result.IfStmt = ifStmt
			result.IfStmtComments = cmap[ifStmt]
		}

		if fileset.Position(current.End()).Line != lineNum {
			return true // not yet at call expression
		}

		if len(result.CallComments) == 0 {
			result.CallComments = cmap[current]
		}

		ce, ok := current.(*ast.CallExpr)
		if !ok {
			return true // not yet at call expression
		}

		// TODO: use type system instead of name match
		se, ok := ce.Fun.(*ast.SelectorExpr)
		if !ok {
			return true // not a call to vt.Message
		}

		switch se.Sel.Name {
		case "Got", "GotWant":
		default:
			return true
		}
		x, ok := se.X.(*ast.Ident)
		if !ok {
			return true
		}
		if x.Name != "vt" {
			return true
		}

		result.CallExpr = ce
		return false
	}
	ast.Inspect(src.AST, pre)
	return result
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
	out, err := formatNode(n.Node)
	if err != nil {
		return fmt.Sprintf("failed to format %s: %s", n.Node, err)
	}
	return fmt.Sprintf("(%T) %s", n.Node, out)
}

// formatNode formats the node using go/format.Node and return the result as a string
func formatNode(node ast.Node) (string, error) {
	buf := new(bytes.Buffer)
	err := format.Node(buf, token.NewFileSet(), node)
	return buf.String(), err
}
