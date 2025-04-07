package parser

import (
	"fmt"
	"strconv"

	"github.com/carlosrabelo/pbasic/pbasic/internal/ast"
	"github.com/carlosrabelo/pbasic/pbasic/internal/lexer"
	"github.com/carlosrabelo/pbasic/pbasic/internal/token"
)

type (
	prefixFn func() ast.Expression
	infixFn  func(ast.Expression) ast.Expression
)

type Parser struct {
	l      *lexer.Lexer
	errors []string

	cur  token.Token
	peek token.Token

	prefixFns map[token.Type]prefixFn
	infixFns  map[token.Type]infixFn
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l}
	p.prefixFns = make(map[token.Type]prefixFn)
	p.infixFns = make(map[token.Type]infixFn)

	p.registerPrefix(token.NUMBER, p.parseNumber)
	p.registerPrefix(token.STRING, p.parseString)
	p.registerPrefix(token.IDENT, p.parseIdent)
	p.registerPrefix(token.MINUS, p.parseUnary)
	p.registerPrefix(token.LPAREN, p.parseGrouped)
	p.registerPrefix(token.FREE, p.parseFree)
	p.registerPrefix(token.RND, p.parseRnd)
	p.registerPrefix(token.ABS, p.parseAbs)

	p.registerInfix(token.PLUS, p.parseInfix)
	p.registerInfix(token.MINUS, p.parseInfix)
	p.registerInfix(token.ASTERISK, p.parseInfix)
	p.registerInfix(token.SLASH, p.parseInfix)
	p.registerInfix(token.ASSIGN, p.parseInfix)
	p.registerInfix(token.NEQ, p.parseInfix)
	p.registerInfix(token.LT, p.parseInfix)
	p.registerInfix(token.GT, p.parseInfix)
	p.registerInfix(token.LE, p.parseInfix)
	p.registerInfix(token.GE, p.parseInfix)

	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) nextToken() {
	p.cur = p.peek
	p.peek = p.l.NextToken()
}

func (p *Parser) curIs(t token.Type) bool {
	return p.cur.Type == t
}

func (p *Parser) curIn(types ...token.Type) bool {
	for _, t := range types {
		if p.cur.Type == t {
			return true
		}
	}
	return false
}

func (p *Parser) peekIs(t token.Type) bool {
	return p.peek.Type == t
}

func (p *Parser) expectPeek(t token.Type) bool {
	if p.peekIs(t) {
		p.nextToken()
		return true
	}
	p.expectError(t)
	return false
}

func (p *Parser) expectError(t token.Type) {
	p.errors = append(p.errors,
		fmt.Sprintf("expected %s, got %s at %s", t, p.peek.Type, p.peek.Pos))
}

func (p *Parser) Errors() []string {
	return p.errors
}

const (
	_ int = iota
	LOWEST
	COND
	SUM
	PRODUCT
	PREFIX
	CALL
)

var precedences = map[token.Type]int{
	token.ASSIGN:   COND,
	token.NEQ:      COND,
	token.LT:       COND,
	token.GT:       COND,
	token.LE:       COND,
	token.GE:       COND,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.ASTERISK: PRODUCT,
	token.SLASH:    PRODUCT,
}

func (p *Parser) peekPrecedence() int {
	if prec, ok := precedences[p.peek.Type]; ok {
		return prec
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if prec, ok := precedences[p.cur.Type]; ok {
		return prec
	}
	return LOWEST
}

func (p *Parser) registerPrefix(t token.Type, fn prefixFn) {
	p.prefixFns[t] = fn
}

func (p *Parser) registerInfix(t token.Type, fn infixFn) {
	p.infixFns[t] = fn
}

// ParseLine parses a complete line of BASIC, returning a Statement
// (which may be a BlockStmt if colon-separated).
func (p *Parser) ParseLine() ast.Statement {
	if p.curIs(token.EOF) {
		return nil
	}
	if p.curIs(token.EOL) {
		p.nextToken()
		return nil
	}

	stmts := p.parseStatementList(token.EOL, token.EOF)
	if p.curIs(token.EOL) {
		p.nextToken()
	}
	if len(stmts) == 0 {
		return nil
	}
	if len(stmts) == 1 {
		return stmts[0]
	}
	return &ast.BlockStmt{Statements: stmts}
}

// parseStatementList parses statements separated by COLON until one of ends.
func (p *Parser) parseStatementList(ends ...token.Type) []ast.Statement {
	var stmts []ast.Statement
	for !p.curIn(ends...) && !p.curIs(token.EOF) {
		if p.curIs(token.EOL) {
			p.nextToken()
			continue
		}
		if p.curIs(token.COLON) {
			p.nextToken()
			continue
		}
		stmt := p.parseStatement()
		if stmt != nil {
			stmts = append(stmts, stmt)
		}
	}
	return stmts
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.cur.Type {
	case token.LET:
		return p.parseLetStmt()
	case token.IDENT:
		return p.parseImplicitLet()
	default:
		p.errors = append(p.errors,
			fmt.Sprintf("unexpected token %s at %s", p.cur.Type, p.cur.Pos))
		for !p.curIs(token.EOL) && !p.curIs(token.EOF) && !p.curIs(token.COLON) {
			p.nextToken()
		}
		return nil
	}
}

// ---- LET ----

func (p *Parser) parseLetStmt() *ast.LetStmt {
	stmt := &ast.LetStmt{Token: p.cur}
	p.nextToken()
	return p.parseLetAssignment(stmt)
}

func (p *Parser) parseImplicitLet() *ast.LetStmt {
	stmt := &ast.LetStmt{}
	return p.parseLetAssignment(stmt)
}

func (p *Parser) parseLetAssignment(stmt *ast.LetStmt) *ast.LetStmt {
	if !p.curIs(token.IDENT) {
		p.errors = append(p.errors,
			fmt.Sprintf("expected variable at %s", p.cur.Pos))
		return stmt
	}
	stmt.Name = &ast.IdentExpr{Token: p.cur, Name: p.cur.Literal}
	p.nextToken()

	if !p.curIs(token.ASSIGN) {
		p.errors = append(p.errors,
			fmt.Sprintf("expected = at %s", p.cur.Pos))
		return stmt
	}
	p.nextToken()

	stmt.Value = p.parseExpression(LOWEST)
	return stmt
}

// ---- expressions (Pratt parser) ----

func (p *Parser) parseExpression(prec int) ast.Expression {
	prefix := p.prefixFns[p.cur.Type]
	if prefix == nil {
		p.errors = append(p.errors,
			fmt.Sprintf("unexpected token %s at %s", p.cur.Type, p.cur.Pos))
		return nil
	}
	left := prefix()

	for prec < p.curPrecedence() &&
		!p.curIs(token.EOL) && !p.curIs(token.EOF) &&
		!p.curIs(token.COMMA) && !p.curIs(token.SEMICOLON) &&
		!p.curIs(token.COLON) && !p.curIs(token.THEN) &&
		!p.curIs(token.RPAREN) {

		infix := p.infixFns[p.cur.Type]
		if infix == nil {
			return left
		}
		left = infix(left)
	}

	return left
}

func (p *Parser) parseNumber() ast.Expression {
	val, err := strconv.ParseFloat(p.cur.Literal, 64)
	if err != nil {
		p.errors = append(p.errors,
			fmt.Sprintf("invalid number %s at %s", p.cur.Literal, p.cur.Pos))
	}
	expr := &ast.NumberExpr{Token: p.cur, Value: val}
	p.nextToken()
	return expr
}

func (p *Parser) parseString() ast.Expression {
	expr := &ast.StringExpr{Token: p.cur, Value: p.cur.Literal}
	p.nextToken()
	return expr
}

func (p *Parser) parseIdent() ast.Expression {
	expr := &ast.IdentExpr{Token: p.cur, Name: p.cur.Literal}
	p.nextToken()
	return expr
}

func (p *Parser) parseUnary() ast.Expression {
	tok := p.cur
	p.nextToken()
	right := p.parseExpression(PREFIX)
	return &ast.UnaryExpr{Token: tok, Op: tok.Type, Right: right}
}

func (p *Parser) parseGrouped() ast.Expression {
	p.nextToken()
	expr := p.parseExpression(LOWEST)
	if !p.curIs(token.RPAREN) {
		p.errors = append(p.errors,
			fmt.Sprintf("expected ) at %s", p.cur.Pos))
		return expr
	}
	p.nextToken()
	return expr
}

func (p *Parser) parseInfix(left ast.Expression) ast.Expression {
	tok := p.cur
	prec := p.curPrecedence()
	p.nextToken()
	right := p.parseExpression(prec)
	return &ast.BinaryExpr{
		Token: tok,
		Left:  left,
		Op:    tok.Type,
		Right: right,
	}
}

func (p *Parser) parseFree() ast.Expression {
	expr := &ast.FreeExpr{Token: p.cur}
	p.nextToken()
	return expr
}

func (p *Parser) parseRnd() ast.Expression {
	tok := p.cur
	p.nextToken()
	if !p.curIs(token.LPAREN) {
		p.errors = append(p.errors,
			fmt.Sprintf("expected ( at %s", p.cur.Pos))
		return &ast.FuncExpr{Token: tok, Name: "RND"}
	}
	p.nextToken()
	arg := p.parseExpression(LOWEST)
	if !p.curIs(token.RPAREN) {
		p.errors = append(p.errors,
			fmt.Sprintf("expected ) at %s", p.cur.Pos))
		return &ast.FuncExpr{Token: tok, Name: "RND", Arg: arg}
	}
	p.nextToken()
	return &ast.FuncExpr{Token: tok, Name: "RND", Arg: arg}
}

func (p *Parser) parseAbs() ast.Expression {
	tok := p.cur
	p.nextToken()
	if !p.curIs(token.LPAREN) {
		p.errors = append(p.errors,
			fmt.Sprintf("expected ( at %s", p.cur.Pos))
		return &ast.FuncExpr{Token: tok, Name: "ABS"}
	}
	p.nextToken()
	arg := p.parseExpression(LOWEST)
	if !p.curIs(token.RPAREN) {
		p.errors = append(p.errors,
			fmt.Sprintf("expected ) at %s", p.cur.Pos))
		return &ast.FuncExpr{Token: tok, Name: "ABS", Arg: arg}
	}
	p.nextToken()
	return &ast.FuncExpr{Token: tok, Name: "ABS", Arg: arg}
}
