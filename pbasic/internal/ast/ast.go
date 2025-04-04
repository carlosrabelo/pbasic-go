package ast

import (
	"fmt"
	"strings"

	"github.com/carlosrabelo/pbasic/pbasic/internal/token"
)

type Node interface {
	Pos() token.Position
	TokenLiteral() string
	String() string
}

type Statement interface {
	Node
	stmtNode()
}

type Expression interface {
	Node
	exprNode()
}

// Program — list of numbered lines
type Program struct {
	Lines []*LineNode
}

func (p *Program) Pos() token.Position {
	if len(p.Lines) > 0 {
		return p.Lines[0].Pos()
	}
	return token.Position{}
}

func (p *Program) TokenLiteral() string {
	if len(p.Lines) > 0 {
		return p.Lines[0].TokenLiteral()
	}
	return ""
}

func (p *Program) String() string {
	var parts []string
	for _, l := range p.Lines {
		parts = append(parts, l.String())
	}
	return strings.Join(parts, "\n")
}

// LineNode — numbered program line
type LineNode struct {
	Number    int
	Statement Statement
	Tok       token.Token
}

func (ln *LineNode) Pos() token.Position    { return ln.Tok.Pos }
func (ln *LineNode) TokenLiteral() string    { return ln.Tok.Literal }
func (ln *LineNode) String() string          { return fmt.Sprintf("%d %s", ln.Number, ln.Statement) }

// BlockStmt — multiple statements on one line (separated by :)
type BlockStmt struct {
	Statements []Statement
}

func (s *BlockStmt) stmtNode()            {}
func (s *BlockStmt) Pos() token.Position  { return s.Statements[0].Pos() }
func (s *BlockStmt) TokenLiteral() string { return s.Statements[0].TokenLiteral() }
func (s *BlockStmt) String() string {
	var parts []string
	for _, stmt := range s.Statements {
		parts = append(parts, stmt.String())
	}
	return strings.Join(parts, " : ")
}

// LetStmt — LET var = expr
type LetStmt struct {
	Token token.Token
	Name  *IdentExpr
	Value Expression
}

func (s *LetStmt) stmtNode()            {}
func (s *LetStmt) Pos() token.Position  { return s.Token.Pos }
func (s *LetStmt) TokenLiteral() string { return s.Token.Literal }
func (s *LetStmt) String() string       { return fmt.Sprintf("LET %s = %s", s.Name, s.Value) }

// GotoStmt — GOTO expr
type GotoStmt struct {
	Token  token.Token
	Target Expression
}

func (s *GotoStmt) stmtNode()            {}
func (s *GotoStmt) Pos() token.Position  { return s.Token.Pos }
func (s *GotoStmt) TokenLiteral() string { return s.Token.Literal }
func (s *GotoStmt) String() string       { return fmt.Sprintf("GOTO %s", s.Target) }

// GosubStmt — GOSUB expr
type GosubStmt struct {
	Token  token.Token
	Target Expression
}

func (s *GosubStmt) stmtNode()            {}
func (s *GosubStmt) Pos() token.Position  { return s.Token.Pos }
func (s *GosubStmt) TokenLiteral() string { return s.Token.Literal }
func (s *GosubStmt) String() string       { return fmt.Sprintf("GOSUB %s", s.Target) }

// ReturnStmt — RETURN
type ReturnStmt struct {
	Token token.Token
}

func (s *ReturnStmt) stmtNode()            {}
func (s *ReturnStmt) Pos() token.Position  { return s.Token.Pos }
func (s *ReturnStmt) TokenLiteral() string { return s.Token.Literal }
func (s *ReturnStmt) String() string       { return "RETURN" }

// EndStmt — END
type EndStmt struct {
	Token token.Token
}

func (s *EndStmt) stmtNode()            {}
func (s *EndStmt) Pos() token.Position  { return s.Token.Pos }
func (s *EndStmt) TokenLiteral() string { return s.Token.Literal }
func (s *EndStmt) String() string       { return "END" }

// PrintStmt — PRINT item [sep item] ...
type PrintItemKind int

const (
	PrintExpr PrintItemKind = iota
	PrintStr
	PrintTab
	PrintSemic
)

type PrintItem struct {
	Kind PrintItemKind
	Expr Expression
	Str  string
}

type PrintStmt struct {
	Token token.Token
	Items []PrintItem
}

func (s *PrintStmt) stmtNode()            {}
func (s *PrintStmt) Pos() token.Position  { return s.Token.Pos }
func (s *PrintStmt) TokenLiteral() string { return s.Token.Literal }

func (s *PrintStmt) String() string {
	var parts []string
	for _, item := range s.Items {
		switch item.Kind {
		case PrintExpr:
			parts = append(parts, item.Expr.String())
		case PrintStr:
			parts = append(parts, fmt.Sprintf("%q", item.Str))
		case PrintTab:
			parts = append(parts, ",")
		case PrintSemic:
			parts = append(parts, ";")
		}
	}
	return fmt.Sprintf("PRINT %s", strings.Join(parts, " "))
}

// IfStmt — IF expr relop expr THEN stmt
type IfStmt struct {
	Token    token.Token
	Cond     Expression
	ThenStmt Statement
}

func (s *IfStmt) stmtNode()            {}
func (s *IfStmt) Pos() token.Position  { return s.Token.Pos }
func (s *IfStmt) TokenLiteral() string { return s.Token.Literal }

func (s *IfStmt) String() string {
	return fmt.Sprintf("IF %s THEN %s", s.Cond, s.ThenStmt)
}

// InputStmt — INPUT ["prompt" ;] var
type InputStmt struct {
	Token  token.Token
	Prompt string
	Var    *IdentExpr
}

func (s *InputStmt) stmtNode()            {}
func (s *InputStmt) Pos() token.Position  { return s.Token.Pos }
func (s *InputStmt) TokenLiteral() string { return s.Token.Literal }

func (s *InputStmt) String() string {
	if s.Prompt != "" {
		return fmt.Sprintf("INPUT %q; %s", s.Prompt, s.Var)
	}
	return fmt.Sprintf("INPUT %s", s.Var)
}

// RemStmt — REM (comment, rest of line ignored)
type RemStmt struct {
	Token token.Token
	Text  string
}

func (s *RemStmt) stmtNode()            {}
func (s *RemStmt) Pos() token.Position  { return s.Token.Pos }
func (s *RemStmt) TokenLiteral() string { return s.Token.Literal }
func (s *RemStmt) String() string       { return fmt.Sprintf("REM%s", s.Text) }

// ListStmt — LIST
type ListStmt struct {
	Token token.Token
}

func (s *ListStmt) stmtNode()            {}
func (s *ListStmt) Pos() token.Position  { return s.Token.Pos }
func (s *ListStmt) TokenLiteral() string { return s.Token.Literal }
func (s *ListStmt) String() string       { return "LIST" }

// RunStmt — RUN
type RunStmt struct {
	Token token.Token
}

func (s *RunStmt) stmtNode()            {}
func (s *RunStmt) Pos() token.Position  { return s.Token.Pos }
func (s *RunStmt) TokenLiteral() string { return s.Token.Literal }
func (s *RunStmt) String() string       { return "RUN" }

// NewStmt — NEW
type NewStmt struct {
	Token token.Token
}

func (s *NewStmt) stmtNode()            {}
func (s *NewStmt) Pos() token.Position  { return s.Token.Pos }
func (s *NewStmt) TokenLiteral() string { return s.Token.Literal }
func (s *NewStmt) String() string       { return "NEW" }

// ExitStmt — EXIT
type ExitStmt struct {
	Token token.Token
}

func (s *ExitStmt) stmtNode()            {}
func (s *ExitStmt) Pos() token.Position  { return s.Token.Pos }
func (s *ExitStmt) TokenLiteral() string { return s.Token.Literal }
func (s *ExitStmt) String() string       { return "EXIT" }

// FreeExpr — FREE function (no arguments)
type FreeExpr struct {
	Token token.Token
}

func (e *FreeExpr) exprNode()            {}
func (e *FreeExpr) Pos() token.Position  { return e.Token.Pos }
func (e *FreeExpr) TokenLiteral() string { return e.Token.Literal }
func (e *FreeExpr) String() string       { return "FREE" }

// Expressions

type NumberExpr struct {
	Token token.Token
	Value float64
}

func (e *NumberExpr) exprNode()            {}
func (e *NumberExpr) Pos() token.Position  { return e.Token.Pos }
func (e *NumberExpr) TokenLiteral() string { return e.Token.Literal }
func (e *NumberExpr) String() string       { return e.Token.Literal }

type StringExpr struct {
	Token token.Token
	Value string
}

func (e *StringExpr) exprNode()            {}
func (e *StringExpr) Pos() token.Position  { return e.Token.Pos }
func (e *StringExpr) TokenLiteral() string { return e.Token.Literal }
func (e *StringExpr) String() string       { return fmt.Sprintf("%q", e.Value) }

type IdentExpr struct {
	Token token.Token
	Name  string
}

func (e *IdentExpr) exprNode()            {}
func (e *IdentExpr) Pos() token.Position  { return e.Token.Pos }
func (e *IdentExpr) TokenLiteral() string { return e.Token.Literal }
func (e *IdentExpr) String() string       { return e.Name }

type UnaryExpr struct {
	Token token.Token
	Op    token.Type
	Right Expression
}

func (e *UnaryExpr) exprNode()            {}
func (e *UnaryExpr) Pos() token.Position  { return e.Token.Pos }
func (e *UnaryExpr) TokenLiteral() string { return e.Token.Literal }
func (e *UnaryExpr) String() string       { return fmt.Sprintf("(%s%s)", e.Op, e.Right) }

type BinaryExpr struct {
	Token token.Token
	Left  Expression
	Op    token.Type
	Right Expression
}

func (e *BinaryExpr) exprNode()            {}
func (e *BinaryExpr) Pos() token.Position  { return e.Token.Pos }
func (e *BinaryExpr) TokenLiteral() string { return e.Token.Literal }
func (e *BinaryExpr) String() string {
	return fmt.Sprintf("(%s %s %s)", e.Left, e.Op, e.Right)
}

type FuncExpr struct {
	Token token.Token
	Name  string
	Arg   Expression
}

func (e *FuncExpr) exprNode()            {}
func (e *FuncExpr) Pos() token.Position  { return e.Token.Pos }
func (e *FuncExpr) TokenLiteral() string { return e.Token.Literal }
func (e *FuncExpr) String() string       { return fmt.Sprintf("%s(%s)", e.Name, e.Arg) }
