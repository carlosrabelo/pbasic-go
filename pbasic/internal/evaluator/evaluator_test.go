package evaluator

import (
	"fmt"
	"strings"
	"testing"

	"github.com/carlosrabelo/pbasic/pbasic/internal/ast"
	"github.com/carlosrabelo/pbasic/pbasic/internal/lexer"
	"github.com/carlosrabelo/pbasic/pbasic/internal/object"
	"github.com/carlosrabelo/pbasic/pbasic/internal/parser"
)

type testProgramStore struct {
	lines []ast.LineNode
}

func (ps *testProgramStore) Find(num int) (*ast.LineNode, bool) {
	for i := range ps.lines {
		if ps.lines[i].Number == num {
			return &ps.lines[i], true
		}
	}
	return nil, false
}

func (ps *testProgramStore) Lines() []ast.LineNode { return ps.lines }
func (ps *testProgramStore) Len() int              { return len(ps.lines) }
func (ps *testProgramStore) FreeMem() int          { return 1000 }
func (ps *testProgramStore) Clear()                { ps.lines = nil }

type testReader struct {
	data   string
	offset int
}

func (r *testReader) ReadString(delim byte) (string, error) {
	if r.offset >= len(r.data) {
		return "", fmt.Errorf("closed")
	}
	end := strings.IndexByte(r.data[r.offset:], delim)
	if end < 0 {
		s := r.data[r.offset:]
		r.offset = len(r.data)
		return s, nil
	}
	s := r.data[r.offset : r.offset+end+1]
	r.offset += end + 1
	return s, nil
}

func newTestCtx(prog *testProgramStore, input string) *EvalContext {
	var buf strings.Builder
	return &EvalContext{
		Env:   object.NewEnvironment(),
		Prog:  prog,
		Wr:    &buf,
		Rd:    &testReader{data: input},
	}

}

func parse(t *testing.T, input string) ast.Statement {
	t.Helper()
	l := lexer.New(input)
	p := parser.New(l)
	stmt := p.ParseLine()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}
	return stmt
}

func testEval(t *testing.T, input string) (*EvalContext, object.Object) {
	t.Helper()
	stmt := parse(t, input)
	prog := &testProgramStore{}
	ctx := newTestCtx(prog, "42\n")
	result := Eval(stmt, ctx)
	return ctx, result
}

func TestEvalNumberExpr(t *testing.T) {
	ctx, result := testEval(t, "LET A = 42")
	if isError(result) {
		t.Fatalf("unexpected error: %s", result.(*object.Error).Message)
	}
	obj, ok := ctx.Env.Get("A")
	if !ok {
		t.Fatal("variable A not set")
	}
	num, ok := obj.(*object.Number)
	if !ok {
		t.Fatalf("expected Number, got %T", obj)
	}
	if num.Value != 42 {
		t.Fatalf("expected 42, got %f", num.Value)
	}
}

func TestEvalFloatExpr(t *testing.T) {
	ctx, result := testEval(t, "LET A = 3.14")
	if isError(result) {
		t.Fatalf("unexpected error: %s", result.(*object.Error).Message)
	}
	obj, _ := ctx.Env.Get("A")
	num := obj.(*object.Number)
	if num.Value != 3.14 {
		t.Fatalf("expected 3.14, got %f", num.Value)
	}
}

func TestEvalBinaryArith(t *testing.T) {
	tests := []struct {
		input string
		want  float64
	}{
		{"LET A = 10 + 5", 15},
		{"LET A = 10 - 5", 5},
		{"LET A = 10 * 5", 50},
		{"LET A = 10 / 5", 2},
		{"LET A = 10 + 5 * 2", 20},
		{"LET A = (10 + 5) * 2", 30},
		{"LET A = 3.5 + 1.5", 5},
		{"LET A = 10 / 3", 3.3333333333333335},
	}

	for _, tt := range tests {
		ctx, result := testEval(t, tt.input)
		if isError(result) {
			t.Fatalf("%s: unexpected error: %s", tt.input, result.(*object.Error).Message)
		}
		obj, _ := ctx.Env.Get("A")
		num := obj.(*object.Number)
		if num.Value != tt.want {
			t.Fatalf("%s: expected %f, got %f", tt.input, tt.want, num.Value)
		}
	}
}

func TestEvalRelational(t *testing.T) {
	tests := []struct {
		input string
		want  float64
	}{
		{"LET A = 5 = 5", 1},
		{"LET A = 5 = 6", 0},
		{"LET A = 5 <> 6", 1},
		{"LET A = 5 <> 5", 0},
		{"LET A = 5 < 6", 1},
		{"LET A = 5 < 5", 0},
		{"LET A = 5 > 4", 1},
		{"LET A = 5 > 5", 0},
		{"LET A = 5 <= 5", 1},
		{"LET A = 5 <= 4", 0},
		{"LET A = 5 >= 5", 1},
		{"LET A = 5 >= 6", 0},
	}

	for _, tt := range tests {
		ctx, result := testEval(t, tt.input)
		if isError(result) {
			t.Fatalf("%s: unexpected error: %s", tt.input, result.(*object.Error).Message)
		}
		obj, _ := ctx.Env.Get("A")
		num := obj.(*object.Number)
		if num.Value != tt.want {
			t.Fatalf("%s: expected %f, got %f", tt.input, tt.want, num.Value)
		}
	}
}

func TestEvalUnary(t *testing.T) {
	ctx, result := testEval(t, "LET A = -5")
	if isError(result) {
		t.Fatalf("unexpected error: %s", result.(*object.Error).Message)
	}
	obj, _ := ctx.Env.Get("A")
	num := obj.(*object.Number)
	if num.Value != -5 {
		t.Fatalf("expected -5, got %f", num.Value)
	}
}

func TestEvalDivByZero(t *testing.T) {
	_, result := testEval(t, "LET A = 10 / 0")
	if !isError(result) {
		t.Fatal("expected error for division by zero")
	}
}

func TestEvalAbs(t *testing.T) {
	ctx, result := testEval(t, "LET A = ABS(-10)")
	if isError(result) {
		t.Fatalf("unexpected error: %s", result.(*object.Error).Message)
	}
	obj, _ := ctx.Env.Get("A")
	num := obj.(*object.Number)
	if num.Value != 10 {
		t.Fatalf("expected 10, got %f", num.Value)
	}
}

func TestEvalRnd(t *testing.T) {
	ctx, result := testEval(t, "LET A = RND(10)")
	if isError(result) {
		t.Fatalf("unexpected error: %s", result.(*object.Error).Message)
	}
	obj, _ := ctx.Env.Get("A")
	num := obj.(*object.Number)
	if num.Value < 1 || num.Value > 10 {
		t.Fatalf("RND(10) should be 1-10, got %f", num.Value)
	}
}

func TestEvalFree(t *testing.T) {
	_, result := testEval(t, "LET A = FREE")
	if isError(result) {
		t.Fatalf("unexpected error: %s", result.(*object.Error).Message)
	}
}

func TestEvalIfStmtTrue(t *testing.T) {
	stmt := parse(t, "IF 1 = 1 THEN LET A = 42")
	prog := &testProgramStore{}
	ctx := newTestCtx(prog, "")
	Eval(stmt, ctx)
	obj, ok := ctx.Env.Get("A")
	if !ok {
		t.Fatal("A not set")
	}
	num := obj.(*object.Number)
	if num.Value != 42 {
		t.Fatalf("expected 42, got %f", num.Value)
	}
}

func TestEvalIfStmtFalse(t *testing.T) {
	stmt := parse(t, "IF 0 = 1 THEN LET A = 42")
	ctx := newTestCtx(&testProgramStore{}, "")
	Eval(stmt, ctx)
	_, ok := ctx.Env.Get("A")
	if ok {
		t.Fatal("A should not be set")
	}
}

func TestEvalIfStmtBlock(t *testing.T) {
	input := "IF 1 = 1 THEN LET A = 1 : LET B = 2"
	prog := &testProgramStore{}
	ctx := newTestCtx(prog, "")
	stmt := parse(t, input)
	Eval(stmt, ctx)
	a, _ := ctx.Env.Get("A")
	b, _ := ctx.Env.Get("B")
	if a.(*object.Number).Value != 1 {
		t.Fatalf("expected A=1, got %f", a.(*object.Number).Value)
	}
	if b.(*object.Number).Value != 2 {
		t.Fatalf("expected B=2, got %f", b.(*object.Number).Value)
	}
}

func TestEvalGoto(t *testing.T) {
	prog := &testProgramStore{}
	prog.lines = []ast.LineNode{
		{Number: 10, Statement: parse(t, "LET A = 1")},
		{Number: 20, Statement: parse(t, "LET A = 2")},
	}
	ctx := newTestCtx(prog, "")
	ctx.PC = 10
	ctx.Running = true

	stmt := parse(t, "GOTO 20")
	Eval(stmt, ctx)

	if ctx.PC != 20 {
		t.Fatalf("expected PC=20, got %d", ctx.PC)
	}
	if !ctx.Running {
		t.Fatal("expected running=true")
	}
}

func TestEvalGosubReturn(t *testing.T) {
	prog := &testProgramStore{}
	prog.lines = []ast.LineNode{
		{Number: 100, Statement: parse(t, "LET A = 1")},
		{Number: 200, Statement: parse(t, "RETURN")},
	}
	ctx := newTestCtx(prog, "")
	ctx.PC = 100
	ctx.Running = true

	stmt := parse(t, "GOSUB 200")
	Eval(stmt, ctx)

	if ctx.PC != 200 {
		t.Fatalf("expected PC=200, got %d", ctx.PC)
	}
	if len(ctx.GosubStk) != 1 {
		t.Fatalf("expected stack depth 1, got %d", len(ctx.GosubStk))
	}

	returnStmt := parse(t, "RETURN")
	Eval(returnStmt, ctx)

	if ctx.PC != 100 {
		t.Fatalf("expected PC=100 after return, got %d", ctx.PC)
	}
}

func TestEvalInput(t *testing.T) {
	prog := &testProgramStore{}
	ctx := newTestCtx(prog, "42\n")
	stmt := parse(t, `INPUT X`)
	Eval(stmt, ctx)
	obj, ok := ctx.Env.Get("X")
	if !ok {
		t.Fatal("X not set")
	}
	num := obj.(*object.Number)
	if num.Value != 42 {
		t.Fatalf("expected 42, got %f", num.Value)
	}
}

func TestEvalEnd(t *testing.T) {
	ctx := newTestCtx(&testProgramStore{}, "")
	ctx.Running = true
	stmt := parse(t, "END")
	Eval(stmt, ctx)
	if ctx.Running {
		t.Fatal("expected running=false after END")
	}
}

func TestEvalExit(t *testing.T) {
	ctx := newTestCtx(&testProgramStore{}, "")
	stmt := parse(t, "EXIT")
	Eval(stmt, ctx)
	if !ctx.ShouldExit {
		t.Fatal("expected ShouldExit=true after EXIT")
	}
}

func TestEvalNew(t *testing.T) {
	prog := &testProgramStore{}
	prog.lines = []ast.LineNode{{Number: 10, Statement: parse(t, "PRINT 1")}}
	ctx := newTestCtx(prog, "")
	stmt := parse(t, "NEW")
	Eval(stmt, ctx)
	if prog.Len() != 0 {
		t.Fatal("expected empty program after NEW")
	}
}

func TestEvalBlockStmt(t *testing.T) {
	prog := &testProgramStore{}
	ctx := newTestCtx(prog, "")
	stmt := parse(t, "LET A = 1 : LET B = 2")
	Eval(stmt, ctx)
	a, _ := ctx.Env.Get("A")
	b, _ := ctx.Env.Get("B")
	if a.(*object.Number).Value != 1 {
		t.Fatalf("expected A=1, got %f", a.(*object.Number).Value)
	}
	if b.(*object.Number).Value != 2 {
		t.Fatalf("expected B=2, got %f", b.(*object.Number).Value)
	}
}

func TestEvalPrint(t *testing.T) {
	var buf strings.Builder
	prog := &testProgramStore{}
	ctx := &EvalContext{
		Env:  object.NewEnvironment(),
		Prog: prog,
		Wr:   &buf,
	}
	stmt := parse(t, `PRINT "HELLO"`)
	Eval(stmt, ctx)
	if buf.String() != "HELLO\n" {
		t.Fatalf("expected 'HELLO\\n', got %q", buf.String())
	}
}

func TestEvalVarUndefined(t *testing.T) {
	_, result := testEval(t, "LET A = B")
	if !isError(result) {
		t.Fatal("expected error for undefined variable")
	}
}

func TestEvalRunNoProgram(t *testing.T) {
	prog := &testProgramStore{}
	ctx := newTestCtx(prog, "")
	stmt := parse(t, "RUN")
	result := Eval(stmt, ctx)
	if !isError(result) {
		t.Fatal("expected error RUN with no program")
	}
}

func TestObjectNumberInspect(t *testing.T) {
	tests := []struct {
		val  float64
		want string
	}{
		{42, "42"},
		{3.14, "3.14"},
		{0, "0"},
		{-5, "-5"},
		{1.0, "1"},
	}

	for _, tt := range tests {
		n := &object.Number{Value: tt.val}
		if n.Inspect() != tt.want {
			t.Fatalf("Number(%f).Inspect() = %q, want %q", tt.val, n.Inspect(), tt.want)
		}
	}
}
