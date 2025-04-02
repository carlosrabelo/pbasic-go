package lexer

import (
	"testing"

	"github.com/carlosrabelo/pbasic/pbasic/internal/token"
)

func TestNextToken(t *testing.T) {
	input := `10 LET A = 5 + 3
20 PRINT "HELLO"
30 IF A <> 5 THEN GOTO 100
40 REM comment text
50 GOSUB 200
60 RETURN
70 END
80 INPUT "NAME: "; N
90 LIST
100 RUN
110 NEW
120 EXIT
130 FREE
140 ABS(-5)
150 RND(10)
160 A = 3.14 * 2`

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.NUMBER, "10"},
		{token.LET, "LET"},
		{token.IDENT, "A"},
		{token.ASSIGN, "="},
		{token.NUMBER, "5"},
		{token.PLUS, "+"},
		{token.NUMBER, "3"},
		{token.EOL, "\\n"},

		{token.NUMBER, "20"},
		{token.PRINT, "PRINT"},
		{token.STRING, "HELLO"},
		{token.EOL, "\\n"},

		{token.NUMBER, "30"},
		{token.IF, "IF"},
		{token.IDENT, "A"},
		{token.NEQ, "<>"},
		{token.NUMBER, "5"},
		{token.THEN, "THEN"},
		{token.GOTO, "GOTO"},
		{token.NUMBER, "100"},
		{token.EOL, "\\n"},

		{token.NUMBER, "40"},
		{token.REM, "REM"},
		{token.IDENT, "comment"},
		{token.IDENT, "text"},
		{token.EOL, "\\n"},

		{token.NUMBER, "50"},
		{token.GOSUB, "GOSUB"},
		{token.NUMBER, "200"},
		{token.EOL, "\\n"},

		{token.NUMBER, "60"},
		{token.RETURN, "RETURN"},
		{token.EOL, "\\n"},

		{token.NUMBER, "70"},
		{token.END, "END"},
		{token.EOL, "\\n"},

		{token.NUMBER, "80"},
		{token.INPUT, "INPUT"},
		{token.STRING, "NAME: "},
		{token.SEMICOLON, ";"},
		{token.IDENT, "N"},
		{token.EOL, "\\n"},

		{token.NUMBER, "90"},
		{token.LIST, "LIST"},
		{token.EOL, "\\n"},

		{token.NUMBER, "100"},
		{token.RUN, "RUN"},
		{token.EOL, "\\n"},

		{token.NUMBER, "110"},
		{token.NEW, "NEW"},
		{token.EOL, "\\n"},

		{token.NUMBER, "120"},
		{token.EXIT, "EXIT"},
		{token.EOL, "\\n"},

		{token.NUMBER, "130"},
		{token.FREE, "FREE"},
		{token.EOL, "\\n"},

		{token.NUMBER, "140"},
		{token.ABS, "ABS"},
		{token.LPAREN, "("},
		{token.MINUS, "-"},
		{token.NUMBER, "5"},
		{token.RPAREN, ")"},
		{token.EOL, "\\n"},

		{token.NUMBER, "150"},
		{token.RND, "RND"},
		{token.LPAREN, "("},
		{token.NUMBER, "10"},
		{token.RPAREN, ")"},
		{token.EOL, "\\n"},

		{token.NUMBER, "160"},
		{token.IDENT, "A"},
		{token.ASSIGN, "="},
		{token.NUMBER, "3.14"},
		{token.ASTERISK, "*"},
		{token.NUMBER, "2"},
		{token.EOF, ""},
	}

	l := New(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q (lit=%q)",
				i, tt.expectedType, tok.Type, tok.Literal)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestFloatNumber(t *testing.T) {
	input := "3.14"
	l := New(input)
	tok := l.NextToken()
	if tok.Type != token.NUMBER {
		t.Fatalf("expected NUMBER, got=%s", tok.Type)
	}
	if tok.Literal != "3.14" {
		t.Fatalf("expected 3.14, got=%s", tok.Literal)
	}
}

func TestOperators(t *testing.T) {
	input := "<= >= <> = + - * / , ; : ( )"
	tests := []struct {
		typ token.Type
		lit string
	}{
		{token.LE, "<="},
		{token.GE, ">="},
		{token.NEQ, "<>"},
		{token.ASSIGN, "="},
		{token.PLUS, "+"},
		{token.MINUS, "-"},
		{token.ASTERISK, "*"},
		{token.SLASH, "/"},
		{token.COMMA, ","},
		{token.SEMICOLON, ";"},
		{token.COLON, ":"},
		{token.LPAREN, "("},
		{token.RPAREN, ")"},
		{token.EOF, ""},
	}

	l := New(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.typ {
			t.Fatalf("tests[%d] - type wrong. expected=%q, got=%q", i, tt.typ, tok.Type)
		}
	}
}

func TestIllegalChar(t *testing.T) {
	input := "@"
	l := New(input)
	tok := l.NextToken()
	if tok.Type != token.ILLEGAL {
		t.Fatalf("expected ILLEGAL, got=%s", tok.Type)
	}
}
