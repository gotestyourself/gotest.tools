package result_test

import (
	"go/ast"
	"testing"

	"gotest.tools/v3/internal/result"
)

type basicResult struct{}

func (basicResult) FailureMessage() string {
	return "basic failure"
}

type argsResult struct{}

func (argsResult) FailureMessage(args []ast.Expr) string {
	switch {
	case len(args) != 2:
		return "unexpected args"
	case args[0] == nil:
		return "first arg missing"
	case args[1] != nil:
		return "second arg present"
	default:
		return "filtered"
	}
}

func TestFailureMessage(t *testing.T) {
	tests := []struct {
		name   string
		result interface{}
		args   []ast.Expr
		want   string
	}{
		{
			name:   "basic",
			result: basicResult{},
			want:   "basic failure",
		},
		{
			name:   "with comparison args",
			result: argsResult{},
			args: []ast.Expr{
				&ast.Ident{Name: "actual"},
				&ast.CallExpr{},
			},
			want: "filtered",
		},
		{
			name:   "invalid result",
			result: struct{}{},
			want:   "comparison returned invalid Result type: struct {}",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := result.FailureMessage(tc.result, tc.args)
			if got != tc.want {
				t.Fatalf("FailureMessage() = %q; want %q", got, tc.want)
			}
		})
	}
}

func TestUsesArgs(t *testing.T) {
	tests := []struct {
		name   string
		result interface{}
		want   bool
	}{
		{
			name:   "with comparison args",
			result: argsResult{},
			want:   true,
		},
		{
			name:   "basic",
			result: basicResult{},
			want:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := result.UsesArgs(tc.result)
			if got != tc.want {
				t.Fatalf("UsesArgs() = %v; want %v", got, tc.want)
			}
		})
	}
}
