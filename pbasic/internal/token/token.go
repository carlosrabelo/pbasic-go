package token

import "fmt"

type Type string

const (
	ILLEGAL Type = "ILLEGAL"
	EOF     Type = "EOF"
	EOL     Type = "EOL"

	NUMBER Type = "NUMBER"
	STRING Type = "STRING"
	IDENT  Type = "IDENT"

	LET    Type = "LET"
	GOTO   Type = "GOTO"
	GOSUB  Type = "GOSUB"
	PRINT  Type = "PRINT"
	IF     Type = "IF"
	INPUT  Type = "INPUT"
	RETURN Type = "RETURN"
	END    Type = "END"
	LIST   Type = "LIST"
	RUN    Type = "RUN"
	NEW    Type = "NEW"
	EXIT   Type = "EXIT"
	REM    Type = "REM"
	THEN   Type = "THEN"

	FREE Type = "FREE"
	RND  Type = "RND"
	ABS  Type = "ABS"

	PLUS     Type = "+"
	MINUS    Type = "-"
	ASTERISK Type = "*"
	SLASH    Type = "/"
	ASSIGN   Type = "="

	EQ  Type = "=="
	NEQ Type = "<>"
	LT  Type = "<"
	GT  Type = ">"
	LE  Type = "<="
	GE  Type = ">="

	LPAREN    Type = "("
	RPAREN    Type = ")"
	COMMA     Type = ","
	SEMICOLON Type = ";"
	COLON     Type = ":"
)

var keywords = map[string]Type{
	"LET":    LET,
	"GOTO":   GOTO,
	"GOSUB":  GOSUB,
	"PRINT":  PRINT,
	"IF":     IF,
	"INPUT":  INPUT,
	"RETURN": RETURN,
	"END":    END,
	"LIST":   LIST,
	"RUN":    RUN,
	"NEW":    NEW,
	"EXIT":   EXIT,
	"REM":    REM,
	"THEN":   THEN,
	"FREE":   FREE,
	"RND":    RND,
	"ABS":    ABS,
}

func LookupIdent(ident string) Type {
	if t, ok := keywords[ident]; ok {
		return t
	}
	return IDENT
}

type Position struct {
	Line int
	Col  int
}

func (p Position) String() string {
	return fmt.Sprintf("%d:%d", p.Line, p.Col)
}

type Token struct {
	Type    Type
	Literal string
	Pos     Position
}

func (t Token) String() string {
	if t.Type == NUMBER || t.Type == STRING || t.Type == IDENT {
		return fmt.Sprintf("%s(%s)", t.Type, t.Literal)
	}
	return string(t.Type)
}
