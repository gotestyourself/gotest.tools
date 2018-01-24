package cmp

import (
	"bytes"
	"fmt"
	"go/ast"
	"text/template"

	"github.com/gotestyourself/gotestyourself/internal/source"
)

// TODO: deprecate old (string, bool) result in favor of another interface?

// Result of a Comparison.
type Result interface {
	// Success returns true if the comparison was successful.
	Success() bool
}

type templatedResult struct {
	success  bool
	template string
	data     map[string]interface{}
}

func (r templatedResult) Success() bool {
	return r.success
}

func (r templatedResult) FailureMessage(args []ast.Expr) string {
	msg, err := renderMessage(r, args)
	if err != nil {
		return fmt.Sprintf("failed to render failure message: %s", err)
	}
	return msg
}

// ResultSuccess is a constant which is returned by a ComparisonWithResult to
// indicate success.
var ResultSuccess = templatedResult{success: true}

// TemplatedResultFailure returns a Result with a template string and data which will be
// used to format a failure message.
func TemplatedResultFailure(template string, data map[string]interface{}) Result {
	return templatedResult{template: template, data: data}
}

func renderMessage(result templatedResult, args []ast.Expr) (string, error) {
	tmpl, err := template.New("failure").Funcs(tmplFuncs).Parse(result.template)
	if err != nil {
		return "", err
	}
	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, map[string]interface{}{
		"Data": result.data,
		"Args": args,
	})
	return buf.String(), err
}

var tmplFuncs = template.FuncMap{
	"formatNode": source.FormatNode,
}
