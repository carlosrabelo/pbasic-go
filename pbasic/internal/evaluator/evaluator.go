package evaluator

import (
	"fmt"
	"io"
	"math/rand"

	"github.com/carlosrabelo/pbasic/pbasic/internal/ast"
	"github.com/carlosrabelo/pbasic/pbasic/internal/object"
)

// ProgramStore is the interface the evaluator uses to interact with stored programs.
type ProgramStore interface {
	Find(int) (*ast.LineNode, bool)
	Lines() []ast.LineNode
	Len() int
	FreeMem() int
	Clear()
}

type EvalContext struct {
	Env   *object.Environment
	Prog  ProgramStore
	Wr    io.Writer
	Rd    interface{ ReadString(delim byte) (string, error) }
	Running   bool
	PC        int
	GosubStk  []int
	ShouldExit bool
}

func NewEvalContext(prog ProgramStore, w io.Writer, r interface{ ReadString(byte) (string, error) }) *EvalContext {
	return &EvalContext{
		Env:  object.NewEnvironment(),
		Prog: prog,
		Wr:   w,
		Rd:   r,
	}
}

func Eval(node ast.Node, ctx *EvalContext) object.Object {
	switch n := node.(type) {
	case *ast.Program:
		return evalProgram(n, ctx)
	case *ast.LineNode:
		return Eval(n.Statement, ctx)
	case *ast.BlockStmt:
		return evalBlockStmt(n, ctx)
	case *ast.LetStmt:
		return evalLetStmt(n, ctx)
	case *ast.PrintStmt:
		return evalPrintStmt(n, ctx)
	case *ast.IfStmt:
		return evalIfStmt(n, ctx)
	case *ast.GotoStmt:
		return evalGotoStmt(n, ctx)
	case *ast.GosubStmt:
		return evalGosubStmt(n, ctx)
	case *ast.ReturnStmt:
		return evalReturnStmt(n, ctx)
	case *ast.InputStmt:
		return evalInputStmt(n, ctx)
	case *ast.EndStmt:
		return evalEndStmt(n, ctx)
	case *ast.ListStmt:
		return evalListStmt(n, ctx)
	case *ast.RunStmt:
		return evalRunStmt(n, ctx)
	case *ast.NewStmt:
		return evalNewStmt(n, ctx)
	case *ast.ExitStmt:
		return evalExitStmt(n, ctx)
	case *ast.RemStmt:
		return evalRemStmt(n, ctx)
	case *ast.NumberExpr:
		return &object.Number{Value: n.Value}
	case *ast.StringExpr:
		return &object.String{Value: n.Value}
	case *ast.IdentExpr:
		return evalIdentExpr(n, ctx)
	case *ast.UnaryExpr:
		return evalUnaryExpr(n, ctx)
	case *ast.BinaryExpr:
		return evalBinaryExpr(n, ctx)
	case *ast.FreeExpr:
		return evalFreeExpr(n, ctx)
	case *ast.FuncExpr:
		return evalFuncExpr(n, ctx)
	}
	return nil
}

func isError(obj object.Object) bool {
	return obj != nil && obj.Type() == object.ErrorType
}

func evalProgram(prog *ast.Program, ctx *EvalContext) object.Object {
	for _, line := range prog.Lines {
		result := Eval(line, ctx)
		if isError(result) {
			return result
		}
		if ctx.ShouldExit {
			return nil
		}
	}
	return nil
}

// ---- Block ----

func evalBlockStmt(s *ast.BlockStmt, ctx *EvalContext) object.Object {
	for _, stmt := range s.Statements {
		result := Eval(stmt, ctx)
		if isError(result) {
			return result
		}
		if ctx.ShouldExit {
			return nil
		}
	}
	return nil
}

// ---- LET ----

func evalLetStmt(s *ast.LetStmt, ctx *EvalContext) object.Object {
	val := Eval(s.Value, ctx)
	if isError(val) {
		return val
	}
	if s.Name != nil {
		ctx.Env.Set(s.Name.Name, val)
	}
	return nil
}

// ---- PRINT ----

func evalPrintStmt(s *ast.PrintStmt, ctx *EvalContext) object.Object {
	if len(s.Items) == 0 {
		fmt.Fprintln(ctx.Wr)
		return nil
	}

	noCrlf := false
	for i, item := range s.Items {
		switch item.Kind {
		case ast.PrintExpr:
			val := Eval(item.Expr, ctx)
			if isError(val) {
				return val
			}
			fmt.Fprint(ctx.Wr, val.Inspect())

		case ast.PrintStr:
			fmt.Fprint(ctx.Wr, item.Str)

		case ast.PrintTab:
			fmt.Fprint(ctx.Wr, "        ")

		case ast.PrintSemic:
			if i == len(s.Items)-1 {
				noCrlf = true
			}
		}
	}

	if !noCrlf {
		fmt.Fprintln(ctx.Wr)
	}
	return nil
}

// ---- IF ----

func evalIfStmt(s *ast.IfStmt, ctx *EvalContext) object.Object {
	if s.Cond == nil {
		return nil
	}
	condObj := Eval(s.Cond, ctx)
	if isError(condObj) {
		return condObj
	}

	if s.ThenStmt != nil && isTruthy(condObj) {
		return Eval(s.ThenStmt, ctx)
	}
	return nil
}

func isTruthy(obj object.Object) bool {
	switch o := obj.(type) {
	case *object.Number:
		return o.Value != 0
	case *object.Error:
		return false
	default:
		return true
	}
}

// ---- GOTO ----

func evalGotoStmt(s *ast.GotoStmt, ctx *EvalContext) object.Object {
	target := Eval(s.Target, ctx)
	if isError(target) {
		return target
	}
	targetNum, ok := target.(*object.Number)
	if !ok {
		return object.NewError("GOTO target must be a number")
	}
	lineNum := int(targetNum.Value)

	_, found := ctx.Prog.Find(lineNum)
	if !found {
		return object.NewError("LINE %d NOT FOUND", lineNum)
	}

	ctx.PC = lineNum
	ctx.Running = true
	return nil
}

// ---- GOSUB ----

func evalGosubStmt(s *ast.GosubStmt, ctx *EvalContext) object.Object {
	if len(ctx.GosubStk) >= 32 {
		return object.NewError("GOSUB STACK OVERFLOW")
	}

	target := Eval(s.Target, ctx)
	if isError(target) {
		return target
	}
	targetNum, ok := target.(*object.Number)
	if !ok {
		return object.NewError("GOSUB target must be a number")
	}
	lineNum := int(targetNum.Value)

	_, found := ctx.Prog.Find(lineNum)
	if !found {
		return object.NewError("LINE %d NOT FOUND", lineNum)
	}

	ctx.GosubStk = append(ctx.GosubStk, ctx.PC)
	ctx.PC = lineNum
	ctx.Running = true
	return nil
}

// ---- RETURN ----

func evalReturnStmt(_ *ast.ReturnStmt, ctx *EvalContext) object.Object {
	if len(ctx.GosubStk) == 0 {
		return object.NewError("RETURN WITHOUT GOSUB")
	}
	ctx.PC = ctx.GosubStk[len(ctx.GosubStk)-1]
	ctx.GosubStk = ctx.GosubStk[:len(ctx.GosubStk)-1]
	ctx.Running = true
	return nil
}

// ---- INPUT ----

func evalInputStmt(s *ast.InputStmt, ctx *EvalContext) object.Object {
	if s.Prompt != "" {
		fmt.Fprint(ctx.Wr, s.Prompt)
	}

	if s.Var == nil {
		return object.NewError("INPUT: no variable")
	}

	line, err := ctx.Rd.ReadString('\n')
	if err != nil {
		return object.NewError("INPUT: %v", err)
	}
	line = line[:len(line)-1]
	if len(line) > 0 && line[len(line)-1] == '\r' {
		line = line[:len(line)-1]
	}

	var val float64
	if _, err := fmt.Sscanf(line, "%f", &val); err != nil {
		val = 0
	}
	ctx.Env.Set(s.Var.Name, &object.Number{Value: val})
	return nil
}

// ---- END ----

func evalEndStmt(_ *ast.EndStmt, ctx *EvalContext) object.Object {
	ctx.Running = false
	return nil
}

// ---- LIST ----

func evalListStmt(_ *ast.ListStmt, ctx *EvalContext) object.Object {
	for _, line := range ctx.Prog.Lines() {
		fmt.Fprintf(ctx.Wr, "%d %s\n", line.Number, line.Statement)
	}
	return nil
}

// ---- RUN ----

func evalRunStmt(_ *ast.RunStmt, ctx *EvalContext) object.Object {
	if ctx.Prog.Len() == 0 {
		return object.NewError("NO PROGRAM")
	}
	ctx.Env = object.NewEnvironment()
	ctx.GosubStk = ctx.GosubStk[:0]
	first := ctx.Prog.Lines()[0]
	ctx.PC = first.Number
	ctx.Running = true
	return nil
}

// ---- NEW ----

func evalNewStmt(_ *ast.NewStmt, ctx *EvalContext) object.Object {
	ctx.Prog.Clear()
	ctx.Env = object.NewEnvironment()
	ctx.GosubStk = ctx.GosubStk[:0]
	fmt.Fprintln(ctx.Wr, "OK")
	return nil
}

// ---- EXIT ----

func evalExitStmt(_ *ast.ExitStmt, ctx *EvalContext) object.Object {
	ctx.ShouldExit = true
	return nil
}

// ---- REM ----

func evalRemStmt(_ *ast.RemStmt, _ *EvalContext) object.Object {
	return nil
}

// ---- Expressions ----

func evalIdentExpr(e *ast.IdentExpr, ctx *EvalContext) object.Object {
	val, ok := ctx.Env.Get(e.Name)
	if !ok {
		return object.NewError("UNDEFINED VARIABLE %s", e.Name)
	}
	return val
}

func evalUnaryExpr(e *ast.UnaryExpr, ctx *EvalContext) object.Object {
	right := Eval(e.Right, ctx)
	if isError(right) {
		return right
	}
	rightNum, ok := right.(*object.Number)
	if !ok {
		return object.NewError("TYPE MISMATCH: expected number")
	}
	return &object.Number{Value: -rightNum.Value}
}

func evalBinaryExpr(e *ast.BinaryExpr, ctx *EvalContext) object.Object {
	left := Eval(e.Left, ctx)
	if isError(left) {
		return left
	}
	right := Eval(e.Right, ctx)
	if isError(right) {
		return right
	}

	leftNum, lok := left.(*object.Number)
	rightNum, rok := right.(*object.Number)

	switch e.Op {
	case "=", "==", "<>", "<", ">", "<=", ">=":
		if !lok || !rok {
			return &object.Number{Value: 0}
		}
		var result bool
		switch e.Op {
		case "=", "==":
			result = leftNum.Value == rightNum.Value
		case "<>":
			result = leftNum.Value != rightNum.Value
		case "<":
			result = leftNum.Value < rightNum.Value
		case ">":
			result = leftNum.Value > rightNum.Value
		case "<=":
			result = leftNum.Value <= rightNum.Value
		case ">=":
			result = leftNum.Value >= rightNum.Value
		}
		if result {
			return &object.Number{Value: 1}
		}
		return &object.Number{Value: 0}
	}

	if !lok || !rok {
		return object.NewError("TYPE MISMATCH: expected numbers")
	}

	switch e.Op {
	case "+":
		return &object.Number{Value: leftNum.Value + rightNum.Value}
	case "-":
		return &object.Number{Value: leftNum.Value - rightNum.Value}
	case "*":
		return &object.Number{Value: leftNum.Value * rightNum.Value}
	case "/":
		if rightNum.Value == 0 {
			return object.NewError("DIVISION BY ZERO")
		}
		return &object.Number{Value: leftNum.Value / rightNum.Value}
	}

	return object.NewError("UNKNOWN OPERATOR: %s", e.Op)
}

func evalFreeExpr(_ *ast.FreeExpr, ctx *EvalContext) object.Object {
	return &object.Number{Value: float64(ctx.Prog.FreeMem())}
}

func evalFuncExpr(e *ast.FuncExpr, ctx *EvalContext) object.Object {
	switch e.Name {
	case "RND":
		arg := Eval(e.Arg, ctx)
		if isError(arg) {
			return arg
		}
		argNum, ok := arg.(*object.Number)
		if !ok {
			return object.NewError("RND ARGUMENT MUST BE NUMBER")
		}
		n := int(argNum.Value)
		if n <= 0 {
			return &object.Number{Value: float64(rand.Int())}
		}
		return &object.Number{Value: float64(rand.Intn(n) + 1)}

	case "ABS":
		arg := Eval(e.Arg, ctx)
		if isError(arg) {
			return arg
		}
		argNum, ok := arg.(*object.Number)
		if !ok {
			return object.NewError("ABS ARGUMENT MUST BE NUMBER")
		}
		if argNum.Value < 0 {
			return &object.Number{Value: -argNum.Value}
		}
		return argNum
	}

	return object.NewError("UNKNOWN FUNCTION: %s", e.Name)
}
