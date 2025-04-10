package evaluator

import (
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
