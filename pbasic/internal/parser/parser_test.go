package parser

import (
	"testing"

	"github.com/carlosrabelo/pbasic/pbasic/internal/ast"
	"github.com/carlosrabelo/pbasic/pbasic/internal/lexer"
)

func parse(t *testing.T, input string) ast.Statement {
	t.Helper()
	l := lexer.New(input)
	p := New(l)
	stmt := p.ParseLine()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}
	return stmt
}

func TestLetStmt(t *testing.T) {
	tests := []struct {
		input     string
		name      string
		wantValue float64
	}{
		{"LET A = 10", "A", 10},
		{"LET B = 3.14", "B", 3.14},
		{"LET X = 0", "X", 0},
	}

	for _, tt := range tests {
		stmt := parse(t, tt.input)
		letStmt, ok := stmt.(*ast.LetStmt)
		if !ok {
			t.Fatalf("not a LetStmt: %T", stmt)
		}
		if letStmt.Name.Name != tt.name {
			t.Fatalf("variable name wrong. expected=%q, got=%q", tt.name, letStmt.Name.Name)
		}
		numExpr, ok := letStmt.Value.(*ast.NumberExpr)
		if !ok {
			t.Fatalf("value not a NumberExpr: %T", letStmt.Value)
		}
		if numExpr.Value != tt.wantValue {
			t.Fatalf("value wrong. expected=%f, got=%f", tt.wantValue, numExpr.Value)
		}
	}
}

func TestImplicitLet(t *testing.T) {
	stmt := parse(t, "A = 5")
	letStmt, ok := stmt.(*ast.LetStmt)
	if !ok {
		t.Fatalf("not a LetStmt: %T", stmt)
	}
	if letStmt.Name.Name != "A" {
		t.Fatalf("expected A, got %s", letStmt.Name.Name)
	}
}

func TestExpressionPrecedence(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"A + B + C", "((A + B) + C)"},
		{"A + B * C", "(A + (B * C))"},
		{"A * B + C", "((A * B) + C)"},
		{"A * B * C", "((A * B) * C)"},
		{"A + B - C", "((A + B) - C)"},
		{"-A + B", "((-A) + B)"},
		{"A = B", "(A = B)"},
		{"A <> B", "(A <> B)"},
		{"A < B", "(A < B)"},
		{"A <= B", "(A <= B)"},
		{"A > B", "(A > B)"},
		{"A >= B", "(A >= B)"},
		{"ABS(-5)", "ABS((-5))"},
		{"RND(10)", "RND(10)"},
		{"FREE", "FREE"},
		{"(A + B) * C", "((A + B) * C)"},
	}

	for _, tt := range tests {
		stmt := parse(t, "LET X = "+tt.input)
		letStmt, ok := stmt.(*ast.LetStmt)
		if !ok {
			t.Fatalf("not a LetStmt: %T", stmt)
		}
		got := letStmt.Value.String()
		if got != tt.expected {
			t.Fatalf("%s: expected %q, got %q", tt.input, tt.expected, got)
		}
	}
}

func TestNumberLiteral(t *testing.T) {
	stmt := parse(t, "LET A = 3.14")
	letStmt := stmt.(*ast.LetStmt)
	numExpr := letStmt.Value.(*ast.NumberExpr)
	if numExpr.Value != 3.14 {
		t.Fatalf("expected 3.14, got %f", numExpr.Value)
	}
}

func TestParseErrors(t *testing.T) {
	l := lexer.New("LET = 5")
	p := New(l)
	_ = p.ParseLine()
	if len(p.Errors()) == 0 {
		t.Fatal("expected parse errors")
	}
}
