package source

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/ioutil"
	"runtime"
	"strings"

	"github.com/pkg/errors"
)

// maxContextLines is the maximum number of lines to scan for a complete
// expression
const maxContextLines = 20

// GetCondition returns the condition string by reading it from the file
// identified in the callstack. In golang 1.9 the line number changed from
// being the line where the statement ended to the line where the statement began.
func GetCondition(argPos int) (string, error) {
	lines, err := getSourceLines()
	if err != nil {
		return "", err
	}

	for i := range lines {
		node, err := parser.ParseExpr(getSource(lines, i))
		if err == nil {
			return getArgSourceFromAST(node, argPos)
		}
	}
	return "", errors.Wrapf(err, "failed to parse source")
}

// getSourceLines returns the source line which called skip.If() along with a
// few preceding lines. To properly parse the AST a complete statement is
// required, and that statement may be split across multiple lines, so include
// up to maxContextLines.
func getSourceLines() ([]string, error) {
	const stackIndex = 3
	_, filename, lineNum, ok := runtime.Caller(stackIndex)
	if !ok {
		return nil, errors.New("failed to get caller info")
	}

	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read source file: %s", filename)
	}

	lines := strings.Split(string(raw), "\n")
	if len(lines) < lineNum {
		return nil, errors.Errorf("file %s does not have line %d", filename, lineNum)
	}
	firstLine, lastLine := getSourceLinesRange(lineNum, len(lines))
	return lines[firstLine:lastLine], nil
}

func getArgSourceFromAST(node ast.Expr, argPos int) (string, error) {
	switch expr := node.(type) {
	case *ast.CallExpr:
		buf := new(bytes.Buffer)
		err := format.Node(buf, token.NewFileSet(), expr.Args[argPos])
		return buf.String(), err
	}
	return "", errors.New("unexpected ast")
}
