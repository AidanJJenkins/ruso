package parser

import (
	"fmt"

	"github.com/aidanjjenkins/compiler/ast"
	"github.com/aidanjjenkins/compiler/lexer"
	"github.com/aidanjjenkins/compiler/token"
)

type Parser struct {
	l *lexer.Lexer

	curToken  token.Token
	peekToken token.Token
	errors    []string
}

func (p *Parser) Errs() []string {
	return p.errors
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead", t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) parseUpdateStatement() *ast.UpdateStatement {
	stmt := &ast.UpdateStatement{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}
	stmt.TName = &ast.Identifier{Token: p.curToken, Val: p.curToken.Literal}

	if !p.expectPeek(token.SET) {
		return nil
	}

	cols, val, err := p.parseSet()
	if err == false {
		return nil
	}
	stmt.Cols = append(stmt.Cols, cols)
	stmt.Values = append(stmt.Values, val)

	for !p.curTokenIs(token.WHERE) {
		cols, val, err := p.parseSet()
		if err == false {
			return nil
		}
		stmt.Cols = append(stmt.Cols, cols)
		stmt.Values = append(stmt.Values, val)
	}

	if p.curToken.Type != token.WHERE {
		return nil
	}

	cons := []*ast.Condition{}
	c := p.parseCondition()

	cons = append(cons, c)

	for p.curTokenIs(token.AND) {
		c := p.parseCondition()
		cons = append(cons, c)
	}

	stmt.Condition = cons

	for !p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseSet() (*ast.Identifier, string, bool) {
	if !p.expectPeek(token.IDENT) {
		return nil, "", false
	}

	col := &ast.Identifier{Token: p.curToken, Val: p.curToken.Literal}

	if !p.expectPeek(token.ASSIGN) {
		return nil, "", false
	}

	if !p.expectPeek(token.STRING) {
		return nil, "", false
	}

	value := p.curToken.Literal

	p.nextToken()
	return col, value, true
}

func (p *Parser) parseDeleteStatement() *ast.DeleteStatement {
	stmt := &ast.DeleteStatement{Token: p.curToken}

	if !p.expectPeek(token.FROM) {
		return nil
	}

	if !p.expectPeek(token.IDENT) {
		return nil
	}
	stmt.TName = &ast.Identifier{Token: p.curToken, Val: p.curToken.Literal}

	if !p.expectPeek(token.WHERE) {
		return nil
	}
	cons := []*ast.Condition{}
	c := p.parseCondition()

	cons = append(cons, c)

	for p.curTokenIs(token.AND) {
		c := p.parseCondition()
		cons = append(cons, c)
	}

	stmt.Condition = cons

	for !p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseSelectStatement() *ast.SelectStatement {
	stmt := &ast.SelectStatement{Token: p.curToken}

	if !p.expectPeek(token.ALL) {
		return nil
	}

	if !p.expectPeek(token.FROM) {
		return nil
	}

	if !p.expectPeek(token.IDENT) {
		return nil
	}
	stmt.TName = &ast.Identifier{Token: p.curToken, Val: p.curToken.Literal}

	if !p.expectPeek(token.WHERE) {
		return nil
	}
	cons := []*ast.Condition{}
	c := p.parseCondition()

	cons = append(cons, c)

	for p.curTokenIs(token.AND) {
		c := p.parseCondition()
		cons = append(cons, c)
	}

	stmt.Condition = cons

	for !p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseCondition() *ast.Condition {
	cond := &ast.Condition{}
	if !p.expectPeek(token.IDENT) {
		return nil
	}
	cond.CName = &ast.Identifier{Token: p.curToken, Val: p.curToken.Literal}

	if !p.expectPeek(token.ASSIGN) {
		return nil
	}
	if !p.expectPeek(token.STRING) {
		return nil
	}

	cond.CIdent = p.curToken.Literal

	p.nextToken()
	return cond
}

func (p *Parser) parseInsertStatement() *ast.InsertStatement {
	stmt := &ast.InsertStatement{Token: p.curToken}

	if !p.expectPeek(token.INTO) {
		return nil
	}

	if !p.expectPeek(token.IDENT) {
		return nil
	}
	stmt.TName = &ast.Identifier{Token: p.curToken, Val: p.curToken.Literal}

	p.nextToken()
	if p.curTokenIs(token.VALUES) {
		if !p.expectPeek(token.LPAREN) {
			return nil
		}
		toInsert := &ast.InsertVals{}
		if p.curTokenIs(token.LPAREN) {
			toInsert.Token = p.curToken
			list := p.parseValsList()
			toInsert.Values = list
			stmt.Vals = toInsert
		}
	} else if p.curTokenIs(token.LPAREN) {
		cols := &ast.InsertCols{}
		if p.curTokenIs(token.LPAREN) {
			cols.Token = p.curToken
			list := p.parseColsList()
			cols.Values = list
			stmt.Cols = cols
		}

		if !p.expectPeek(token.VALUES) {
			return nil
		}

		if !p.expectPeek(token.LPAREN) {
			return nil
		}

		toInsert := &ast.InsertVals{}
		toInsert.Token = p.curToken
		list := p.parseValsList()
		toInsert.Values = list
		stmt.Vals = toInsert

		if len(stmt.Cols.Values) != len(stmt.Vals.Values) {
			msg := fmt.Sprintf("Mismatching number of columns and values given. Columns: %d, values: %d", len(stmt.Cols.Values), len(stmt.Vals.Values))
			p.errors = append(p.errors, msg)
			return nil
		}
	}

	for !p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// need to figure out how to identify if its a bool, a string, or an int?
func (p *Parser) parseValue() (ast.Statement, bool) {
	switch p.curToken.Type {
	case token.STRING:
		return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}, true
	case token.INT:
		return &ast.IntegerLiteral{Token: p.curToken, Value: p.curToken.Literal}, true
	case token.TRUE, token.FALSE:
		return &ast.BooleanLiteral{Token: p.curToken, Value: p.curToken.Type == token.TRUE}, true
	default:
		return nil, false
	}
}
func (p *Parser) parseIdentifier() (*ast.Ident, bool) {
	startToken := p.curToken
	combinedLiteral := p.curToken.Literal

	for !p.peekTokenIs(token.COMMA) && !p.peekTokenIs(token.RPAREN) && !p.peekTokenIs(token.EOF) {
		p.nextToken()
		combinedLiteral += p.curToken.Literal
	}

	return &ast.Ident{Token: startToken, Val: combinedLiteral}, true
}

func (p *Parser) parseValsList() []ast.Statement {

	values := []ast.Statement{}

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return nil
	}

	p.nextToken()
	value, ok := p.parseValue()
	if !ok {
		return nil
	}
	values = append(values, value)

	// Parse subsequent values
	for p.peekTokenIs(token.COMMA) {
		p.nextToken() // Consume the comma
		p.nextToken() // Move to the next value

		value, ok := p.parseValue()
		if !ok {
			return nil
		}
		values = append(values, value)
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return values
}

func (p *Parser) parseInsertValues() (string, bool) {
	if !p.expectPeek(token.STRING) {
		return "", false
	}
	res := p.curToken.Literal

	if p.curTokenIs(token.RPAREN) {
		return res, true
	}
	p.nextToken()
	return res, true
}

func (p *Parser) parseColsList() []*ast.Ident {
	values := []*ast.Ident{}

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return nil
	}

	p.nextToken()
	value, ok := p.parseIdentifier()
	if !ok {
		return nil
	}
	values = append(values, value)

	// Parse subsequent values
	for p.peekTokenIs(token.COMMA) {
		p.nextToken() // Consume the comma
		p.nextToken() // Move to the next value

		value, ok := p.parseIdentifier()
		if !ok {
			return nil
		}
		values = append(values, value)
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return values
}

func (p *Parser) parseCreateTableStatement() *ast.CreateTableStatement {
	stmt := &ast.CreateTableStatement{Token: p.curToken}

	if !p.expectPeek(token.TABLE) {
		return nil
	}

	if !p.expectPeek(token.IDENT) {
		return nil
	}
	stmt.TName = &ast.Identifier{Token: p.curToken, Val: p.curToken.Literal}

	ok := p.isValidName(stmt.TName.Val)
	if !ok {
		return nil
	}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	c, t, err := p.parseTables()
	if err == false {
		return nil
	}

	stmt.Cols = append(stmt.Cols, c)
	stmt.ColTypes = append(stmt.ColTypes, t)
	for p.curTokenIs(token.COMMA) {
		c, t, err := p.parseTables()
		if err == false {
			return nil
		}

		stmt.Cols = append(stmt.Cols, c)
		stmt.ColTypes = append(stmt.ColTypes, t)
	}

	for !p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseTables() (string, string, bool) {
	if !p.expectPeek(token.IDENT) {
		return "", "", false
	}

	cName := p.curToken.Literal

	if !p.expectPeek(token.IDENT) {
		return "", "", false
	}
	cType := p.curToken.Literal

	if p.curTokenIs(token.RPAREN) {
		return "", "", false
	}
	p.nextToken()
	return cName, cType, true
}

func (p *Parser) parseIndexStatement() *ast.CreateIndexStatement {
	stmt := &ast.CreateIndexStatement{Token: p.curToken}

	if !p.expectPeek(token.INDEX) {
		return nil
	}
	if !p.expectPeek(token.ON) {
		return nil
	}

	if !p.expectPeek(token.IDENT) {
		return nil
	}
	stmt.TName = &ast.Identifier{Token: p.curToken, Val: p.curToken.Literal}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	val := p.parseIndexes()
	if val == nil {
		return nil
	}

	stmt.Cols = append(stmt.Cols, val)
	for p.curTokenIs(token.COMMA) {
		val := p.parseIndexes()
		if val == nil {
			return nil
		}
		stmt.Cols = append(stmt.Cols, val)
	}

	for !p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseIndexes() *ast.Identifier {
	if !p.expectPeek(token.IDENT) {
		return nil
	}

	res := &ast.Identifier{Token: p.curToken, Val: p.curToken.Literal}

	if p.curTokenIs(token.RPAREN) {
		return nil
	}
	p.nextToken()
	return res
}

func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.SELECT:
		return p.parseSelectStatement()
	case token.DELETE:
		return p.parseDeleteStatement()
	case token.INSERT:
		return p.parseInsertStatement()
	case token.UPDATE:
		return p.parseUpdateStatement()
	case token.CREATE:
		if p.peekToken.Type == token.INDEX {
			return p.parseIndexStatement()
		}
		return p.parseCreateTableStatement()
	default:
		return nil
	}
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for p.curToken.Type != token.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}

	return program
}

func (p *Parser) isValidName(name string) bool {
	if len(name) > 255 {
		msg := "name is too long; must be 255 characters or fewer"
		p.errors = append(p.errors, msg)
		return false
	}

	for i := 0; i < len(name); i++ {
		if name[i] == ' ' {
			msg := "name contains spaces"
			p.errors = append(p.errors, msg)
			return false
		}

		// Check for other problematic characters
		switch name[i] {
		case '\'', '"', ';', '\\', '/', '.', ',', '!', '@', '#', '$', '%', '^', '&', '*', '(', ')', '-', '+', '=', '{', '}', '[', ']', ':', '<', '>', '?', '|', '`', '~':
			msg := "name contains invalid characters"
			p.errors = append(p.errors, msg)
			return false
		}
	}

	return true
}
