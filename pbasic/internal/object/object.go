package object

import (
	"fmt"
	"math"

	"github.com/carlosrabelo/pbasic/pbasic/internal/token"
)

type Type string

const (
	NumberType Type = "NUMBER"
	StringType Type = "STRING"
	ErrorType  Type = "ERROR"
)

type Object interface {
	Type() Type
	Inspect() string
}

type Number struct {
	Value float64
}

func (n *Number) Type() Type { return NumberType }
func (n *Number) Inspect() string {
	if n.Value == math.Trunc(n.Value) && !math.IsInf(n.Value, 0) && math.Abs(n.Value) < 1e15 {
		return fmt.Sprintf("%d", int64(n.Value))
	}
	return fmt.Sprintf("%g", n.Value)
}

type String struct {
	Value string
}

func (s *String) Type() Type     { return StringType }
func (s *String) Inspect() string { return s.Value }

type Error struct {
	Message string
	Pos     token.Position
}

func (e *Error) Type() Type { return ErrorType }
func (e *Error) Inspect() string {
	if e.Pos.Line != 0 || e.Pos.Col != 0 {
		return fmt.Sprintf("?ERROR at %s: %s", e.Pos, e.Message)
	}
	return fmt.Sprintf("?%s", e.Message)
}

func NewError(format string, a ...any) *Error {
	return &Error{Message: fmt.Sprintf(format, a...)}
}

func NewErrorPos(pos token.Position, format string, a ...any) *Error {
	return &Error{Message: fmt.Sprintf(format, a...), Pos: pos}
}

// Environment maps variable names to values
type Environment struct {
	store map[string]Object
}

func NewEnvironment() *Environment {
	return &Environment{store: make(map[string]Object)}
}

func (e *Environment) Get(name string) (Object, bool) {
	obj, ok := e.store[name]
	return obj, ok
}

func (e *Environment) Set(name string, val Object) Object {
	e.store[name] = val
	return val
}
