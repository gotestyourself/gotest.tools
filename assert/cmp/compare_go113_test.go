// +build go1.13

package cmp

import (
	"go/ast"
	"os"
	"testing"
)

func TestErrorIs(t *testing.T) {
	t.Run("equal", func(t *testing.T) {
		result := ErrorIs(stubError{}, stubError{})()
		assertSuccess(t, result)
	})
	t.Run("actual is nil", func(t *testing.T) {
		result := ErrorIs(nil, stubError{})()
		args := []ast.Expr{
			&ast.Ident{Name: "err"},
			&ast.StarExpr{
				X: &ast.SelectorExpr{
					X:   &ast.Ident{Name: "errors"},
					Sel: &ast.Ident{Name: "errorString"},
				}},
		}
		expected := `error is nil, not "stub error" (*errors.errorString cmp.stubError)`
		assertFailureTemplate(t, result, args, expected)
	})
	t.Run("not equal", func(t *testing.T) {
		result := ErrorIs(os.ErrClosed, stubError{})()
		args := []ast.Expr{
			&ast.Ident{Name: "err"},
			&ast.StarExpr{
				X: &ast.SelectorExpr{
					X:   &ast.Ident{Name: "errors"},
					Sel: &ast.Ident{Name: "errorString"},
				}},
		}
		expected := `error is "file already closed" (err *errors.errorString), not "stub error" (*errors.errorString cmp.stubError)`
		assertFailureTemplate(t, result, args, expected)
	})
}
