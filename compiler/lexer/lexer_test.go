package lexer

import (
	"testing"

	"github.com/aidanjjenkins/compiler/token"
)

func TestNextToke(t *testing.T) {
	input := `INSERT INTO dogs (id, name) VALUES (1, 'Winnie');

	SELECT * FROM dogs;

	UPDATE dogs SET age = 3 WHERE name = 'Winnie';

	CREATE TABLE dogs (
		name VARCHAR(255) UNIQUE,
		age INT
	);

	DELETE * FROM dogs;
`
	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.INSERT, "INSERT"},
		{token.INTO, "INTO"},
		{token.IDENT, "dogs"},
		{token.LPAREN, "("},
		{token.IDENT, "id"},
		{token.COMMA, ","},
		{token.IDENT, "name"},
		{token.RPAREN, ")"},
		{token.VALUES, "VALUES"},
		{token.LPAREN, "("},
		{token.INTEGER, "1"},
		{token.COMMA, ","},
		{token.SQUOTE, "'"},
		{token.IDENT, "Winnie"},
		{token.SQUOTE, "'"},
		{token.RPAREN, ")"},
		{token.SEMICOLON, ";"},
		{token.SELECT, "SELECT"},
		{token.ALL, "*"},
		{token.FROM, "FROM"},
		{token.IDENT, "dogs"},
		{token.SEMICOLON, ";"},
		{token.UPDATE, "UPDATE"},
		{token.IDENT, "dogs"},
		{token.SET, "SET"},
		{token.IDENT, "age"},
		{token.ASSIGN, "="},
		{token.INTEGER, "3"},
		{token.WHERE, "WHERE"},
		{token.IDENT, "name"},
		{token.ASSIGN, "="},
		{token.SQUOTE, "'"},
		{token.IDENT, "Winnie"},
		{token.SQUOTE, "'"},
		{token.SEMICOLON, ";"},
		{token.CREATE, "CREATE"},
		{token.TABLE, "TABLE"},
		{token.IDENT, "dogs"},
		{token.LPAREN, "("},
		{token.IDENT, "name"},
		{token.VARCHAR, "VARCHAR"},
		{token.LPAREN, "("},
		{token.INTEGER, "255"},
		{token.RPAREN, ")"},
		{token.UNIQUE, "UNIQUE"},
		{token.COMMA, ","},
		{token.IDENT, "age"},
		{token.INT, "INT"},
		{token.RPAREN, ")"},
		{token.SEMICOLON, ";"},
		{token.DELETE, "DELETE"},
		{token.ALL, "*"},
		{token.FROM, "FROM"},
		{token.IDENT, "dogs"},
		{token.SEMICOLON, ";"},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokenType wrong. expected=%q, got=%q", i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}
