package cmp

import (
	"bytes"
	"fmt"
	"go/ast"
	"text/template"

	"github.com/gotestyourself/gotestyourself/internal/source"
)

// Result of a Comparison.
type Result interface {
	Success() bool
}

type result struct {
	success bool
	message string
}

func (r result) Success() bool {
	return r.success
}

func (r result) FailureMessage() string {
	return r.message
}

// ResultSuccess is a constant which is returned by a ComparisonWithResult to
// indicate success.
var ResultSuccess = result{success: true}

// ResultFailure returns a failed Result with a failure message.
func ResultFailure(message string) Result {
	return result{message: message}
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

// ResultFailureTemplate returns a Result with a template string and data which will be
// used to format a failure message. The template may access data from .Data,
// the comparison args as .Args ([]ast.Expr), and the formatNode function for
// formatting the args.
func ResultFailureTemplate(template string, data map[string]interface{}) Result {
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
