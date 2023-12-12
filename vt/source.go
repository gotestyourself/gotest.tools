package vt

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"os"

	"gotest.tools/v3/internal/source"
)

type messageCallSource struct {
	CallExpr *ast.CallExpr
	IfStmt   *ast.IfStmt
}

// TODO: defer support was removed
func getNodeAtLine(src source.FileSource, lineNum int) (messageCallSource, error) {
	result := scanToLine(src.FileSet, src.AST, lineNum)
	if result.CallExpr == nil || result.IfStmt == nil {
		return result, fmt.Errorf("failed to find an expression on line")
	}
	debug("found node: %s", debugFormatNode{result.IfStmt})
	return result, nil
}

func scanToLine(fileset *token.FileSet, node ast.Node, lineNum int) messageCallSource {
	var result messageCallSource

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
		}

		if fileset.Position(current.End()).Line != lineNum {
			return true // not yet at call expression
		}

		ce, ok := current.(*ast.CallExpr)
		if !ok {
			return true // not yet at call expression
		}
		result.CallExpr = ce
		return false
	}
	ast.Inspect(node, pre)
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
