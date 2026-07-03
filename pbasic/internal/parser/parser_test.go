package parser

import (
	"testing"

	"github.com/carlosrabelo/pbasic/pbasic/internal/ast"
	"github.com/carlosrabelo/pbasic/pbasic/internal/lexer"
	"github.com/carlosrabelo/pbasic/pbasic/internal/token"
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

func TestPrintStmt(t *testing.T) {
	stmt := parse(t, `PRINT "HELLO", X;`)
	printStmt, ok := stmt.(*ast.PrintStmt)
	if !ok {
		t.Fatalf("not a PrintStmt: %T", stmt)
	}
	if len(printStmt.Items) != 4 {
		t.Fatalf("expected 4 items, got %d", len(printStmt.Items))
	}
	if printStmt.Items[0].Kind != ast.PrintStr || printStmt.Items[0].Str != "HELLO" {
		t.Fatalf("first item wrong: %+v", printStmt.Items[0])
	}
	if printStmt.Items[1].Kind != ast.PrintTab {
		t.Fatalf("second item expected tab, got %+v", printStmt.Items[1])
	}
	if printStmt.Items[2].Kind != ast.PrintExpr {
		t.Fatalf("third item expected expr, got %+v", printStmt.Items[2])
	}
	if printStmt.Items[3].Kind != ast.PrintSemic {
		t.Fatalf("fourth item expected semicolon, got %+v", printStmt.Items[3])
	}
}

func TestGotoStmt(t *testing.T) {
	stmt := parse(t, "GOTO 100")
	gotoStmt, ok := stmt.(*ast.GotoStmt)
	if !ok {
		t.Fatalf("not a GotoStmt: %T", stmt)
	}
	numExpr, ok := gotoStmt.Target.(*ast.NumberExpr)
	if !ok {
		t.Fatalf("target not a NumberExpr: %T", gotoStmt.Target)
	}
	if numExpr.Value != 100 {
		t.Fatalf("expected 100, got %f", numExpr.Value)
	}
}

func TestGosubStmt(t *testing.T) {
	stmt := parse(t, "GOSUB 200")
	_, ok := stmt.(*ast.GosubStmt)
	if !ok {
		t.Fatalf("not a GosubStmt: %T", stmt)
	}
}

func TestIfStmt(t *testing.T) {
	stmt := parse(t, "IF A = 5 THEN GOTO 100")
	ifStmt, ok := stmt.(*ast.IfStmt)
	if !ok {
		t.Fatalf("not an IfStmt: %T", stmt)
	}
	binExpr, ok := ifStmt.Cond.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("cond not a BinaryExpr: %T", ifStmt.Cond)
	}
	if binExpr.Op != token.ASSIGN {
		t.Fatalf("expected = operator, got %s", binExpr.Op)
	}
	_, ok = ifStmt.ThenStmt.(*ast.GotoStmt)
	if !ok {
		t.Fatalf("then not a GotoStmt: %T", ifStmt.ThenStmt)
	}
}

func TestIfStmtBlock(t *testing.T) {
	stmt := parse(t, "IF A = 1 THEN PRINT A : GOTO 100")
	ifStmt, ok := stmt.(*ast.IfStmt)
	if !ok {
		t.Fatalf("not an IfStmt: %T", stmt)
	}
	block, ok := ifStmt.ThenStmt.(*ast.BlockStmt)
	if !ok {
		t.Fatalf("then not a BlockStmt: %T", ifStmt.ThenStmt)
	}
	if len(block.Statements) != 2 {
		t.Fatalf("expected 2 statements, got %d", len(block.Statements))
	}
}

func TestInputStmt(t *testing.T) {
	stmt := parse(t, `INPUT "NAME: "; N`)
	inputStmt, ok := stmt.(*ast.InputStmt)
	if !ok {
		t.Fatalf("not an InputStmt: %T", stmt)
	}
	if inputStmt.Prompt != "NAME: " {
		t.Fatalf("expected prompt 'NAME: ', got '%s'", inputStmt.Prompt)
	}
	if inputStmt.Var.Name != "N" {
		t.Fatalf("expected var N, got %s", inputStmt.Var.Name)
	}
}

func TestInputStmtNoPrompt(t *testing.T) {
	stmt := parse(t, "INPUT X")
	inputStmt, ok := stmt.(*ast.InputStmt)
	if !ok {
		t.Fatalf("not an InputStmt: %T", stmt)
	}
	if inputStmt.Prompt != "" {
		t.Fatalf("expected no prompt, got '%s'", inputStmt.Prompt)
	}
	if inputStmt.Var.Name != "X" {
		t.Fatalf("expected X, got %s", inputStmt.Var.Name)
	}
}

func TestRemStmt(t *testing.T) {
	stmt := parse(t, "REM HELLO WORLD")
	remStmt, ok := stmt.(*ast.RemStmt)
	if !ok {
		t.Fatalf("not a RemStmt: %T", stmt)
	}
	if remStmt.Text == "" {
		t.Fatalf("expected non-empty comment text")
	}
}

func TestBlockStmt(t *testing.T) {
	stmt := parse(t, "PRINT A : GOTO 100")
	block, ok := stmt.(*ast.BlockStmt)
	if !ok {
		t.Fatalf("not a BlockStmt: %T", stmt)
	}
	if len(block.Statements) != 2 {
		t.Fatalf("expected 2 statements, got %d", len(block.Statements))
	}
	_, ok = block.Statements[0].(*ast.PrintStmt)
	if !ok {
		t.Fatalf("first not PrintStmt: %T", block.Statements[0])
	}
	_, ok = block.Statements[1].(*ast.GotoStmt)
	if !ok {
		t.Fatalf("second not GotoStmt: %T", block.Statements[1])
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
