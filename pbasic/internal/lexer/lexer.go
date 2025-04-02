package lexer

import (
	"strings"

	"github.com/carlosrabelo/pbasic/pbasic/internal/token"
)

type Lexer struct {
	input   string
	pos     int
	readPos int
	ch      byte
	line    int
	col     int
}

func New(input string) *Lexer {
	l := &Lexer{input: input, line: 1}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPos >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPos]
	}
	l.pos = l.readPos
	l.readPos++
	l.col++
}

func (l *Lexer) peekChar() byte {
	if l.readPos >= len(l.input) {
		return 0
	}
	return l.input[l.readPos]
}

func (l *Lexer) curPos() token.Position {
	return token.Position{Line: l.line, Col: l.col}
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\r' {
		if l.ch == '\r' {
			l.col = 0
		}
		l.readChar()
	}
}

func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	l.skipWhitespace()

	switch l.ch {
	case 0:
		tok = token.Token{Type: token.EOF, Pos: l.curPos()}
	case '\n':
		tok = token.Token{Type: token.EOL, Literal: "\\n", Pos: l.curPos()}
		l.line++
		l.col = 0
		l.readChar()
		return tok
	case '+':
		tok = token.Token{Type: token.PLUS, Literal: "+", Pos: l.curPos()}
		l.readChar()
	case '-':
		tok = token.Token{Type: token.MINUS, Literal: "-", Pos: l.curPos()}
		l.readChar()
	case '*':
		tok = token.Token{Type: token.ASTERISK, Literal: "*", Pos: l.curPos()}
		l.readChar()
	case '/':
		tok = token.Token{Type: token.SLASH, Literal: "/", Pos: l.curPos()}
		l.readChar()
	case '(':
		tok = token.Token{Type: token.LPAREN, Literal: "(", Pos: l.curPos()}
		l.readChar()
	case ')':
		tok = token.Token{Type: token.RPAREN, Literal: ")", Pos: l.curPos()}
		l.readChar()
	case ',':
		tok = token.Token{Type: token.COMMA, Literal: ",", Pos: l.curPos()}
		l.readChar()
	case ';':
		tok = token.Token{Type: token.SEMICOLON, Literal: ";", Pos: l.curPos()}
		l.readChar()
	case ':':
		tok = token.Token{Type: token.COLON, Literal: ":", Pos: l.curPos()}
		l.readChar()
	case '=':
		tok = token.Token{Type: token.ASSIGN, Literal: "=", Pos: l.curPos()}
		l.readChar()
	case '<':
		pos := l.curPos()
		if l.peekChar() == '>' {
			tok = token.Token{Type: token.NEQ, Literal: "<>", Pos: pos}
			l.readChar()
			l.readChar()
		} else if l.peekChar() == '=' {
			tok = token.Token{Type: token.LE, Literal: "<=", Pos: pos}
			l.readChar()
			l.readChar()
		} else {
			tok = token.Token{Type: token.LT, Literal: "<", Pos: pos}
			l.readChar()
		}
	case '>':
		pos := l.curPos()
		if l.peekChar() == '=' {
			tok = token.Token{Type: token.GE, Literal: ">=", Pos: pos}
			l.readChar()
			l.readChar()
		} else {
			tok = token.Token{Type: token.GT, Literal: ">", Pos: pos}
			l.readChar()
		}
	case '"':
		tok = l.readString()
	default:
		if isDigit(l.ch) {
			return l.readNumber()
		}
		if isLetter(l.ch) {
			return l.readIdent()
		}
		tok = token.Token{Type: token.ILLEGAL, Literal: string(l.ch), Pos: l.curPos()}
		l.readChar()
	}

	return tok
}

func (l *Lexer) readNumber() token.Token {
	pos := l.curPos()
	var sb strings.Builder
	for isDigit(l.ch) {
		sb.WriteByte(l.ch)
		l.readChar()
	}
	if l.ch == '.' {
		sb.WriteByte('.')
		l.readChar()
		for isDigit(l.ch) {
			sb.WriteByte(l.ch)
			l.readChar()
		}
	}
	return token.Token{
		Type:    token.NUMBER,
		Literal: sb.String(),
		Pos:     pos,
	}
}

func (l *Lexer) readString() token.Token {
	pos := l.curPos()
	l.readChar()
	var sb strings.Builder
	for l.ch != '"' && l.ch != 0 && l.ch != '\n' {
		sb.WriteByte(l.ch)
		l.readChar()
	}
	if l.ch == '"' {
		l.readChar()
	}
	return token.Token{
		Type:    token.STRING,
		Literal: sb.String(),
		Pos:     pos,
	}
}

func (l *Lexer) readIdent() token.Token {
	pos := l.curPos()
	var sb strings.Builder
	for isLetter(l.ch) {
		sb.WriteByte(l.ch)
		l.readChar()
	}
	lit := sb.String()
	return token.Token{
		Type:    token.LookupIdent(lit),
		Literal: lit,
		Pos:     pos,
	}
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func isLetter(ch byte) bool {
	return ch >= 'A' && ch <= 'Z' || ch >= 'a' && ch <= 'z'
}
