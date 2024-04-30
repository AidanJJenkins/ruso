package parser

import (
	"testing"

	"github.com/aidanjjenkins/compiler/ast"
	"github.com/aidanjjenkins/compiler/lexer"
)

func createParseProgram(input string, t *testing.T) *ast.Program {
	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	return program
}

func TestUpdateStatement(t *testing.T) {
	tests := []struct {
		input              string
		expectedIdentifier string
		expectedCols       []string
		expectedVal        []string

		expectedCond    []string
		expectedCondVal []string
	}{
		{"UPDATE dogs SET name = 'stella', breed = 'labradoodle' WHERE name = 'Winnie';", "dogs", []string{"name", "breed"}, []string{"stella", "labradoodle"}, []string{"name"}, []string{"Winnie"}},
		{"UPDATE dogs SET name = 'stella', breed = 'labradoodle', age = 3 WHERE name = 'Winnie';", "dogs", []string{"name", "breed", "age"}, []string{"stella", "labradoodle", "3"}, []string{"name"}, []string{"Winnie"}},
	}

	for _, tt := range tests {
		program := createParseProgram(tt.input, t)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain 1 statements. got=%d", len(program.Statements))
		}

		stmt := program.Statements[0]
		if !testUpdateStatement(t, stmt, tt.expectedIdentifier, tt.expectedCols, tt.expectedVal, tt.expectedCond, tt.expectedCondVal) {
			return
		}
	}
}

func testUpdateStatement(t *testing.T, s ast.Statement, name string, colName, val, condName, condVal []string) bool {
	if s.TokenLiteral() != "UPDATE" {
		t.Errorf("s.TokenLiteral not 'let'. got=%q", s.TokenLiteral())
		return false
	}

	letStmt, ok := s.(*ast.UpdateStatement)
	if !ok {
		t.Errorf("s not *ast.Delete. got=%T", s)
		return false
	}

	if letStmt.TName.Val != name {
		t.Errorf("letStmt.Name.Value not '%s'. got=%s", name, letStmt.TName.Val)
		return false
	}

	for i := range val {
		if letStmt.Cols[i].Val != colName[i] {
			t.Errorf("Where clause CName value expected: '%s'. got=%s", val[i], letStmt.Cols[i])
			return false
		}
		if letStmt.Values[i] != val[i] {
			t.Errorf("Where clause CName value expected: '%s'. got=%s", val[i], letStmt.Values[i])
			return false
		}
	}

	for j := range condVal {
		if letStmt.Condition[j].CName.Val != condName[j] {
			t.Errorf("Where clause CName value expected: '%s'. got=%s", condName[j], letStmt.Condition[j].CName.Val)
			return false
		}
		if letStmt.Condition[j].CIdent.Val != condVal[j] {
			t.Errorf("Where clause CName value expected: '%s'. got=%s", condVal[j], letStmt.Condition[j].CIdent.Val)
			return false
		}
	}

	return true
}

func TestInsertStatement(t *testing.T) {
	tests := []struct {
		input              string
		expectedIdentifier string
		expectedVal        []string
	}{
		{"INSERT INTO dogs ('stella', 'labradoodle');", "dogs", []string{"stella", "labradoodle"}},
		{"INSERT INTO dogs ('stella', 'labradoodle', '7');", "dogs", []string{"stella", "labradoodle", "7"}},
		{"INSERT INTO dogs ('stella', 'labradoodle', '7', 'untrained');", "dogs", []string{"stella", "labradoodle", "7", "untrained"}},
	}

	for _, tt := range tests {
		program := createParseProgram(tt.input, t)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain 1 statements. got=%d", len(program.Statements))
		}

		stmt := program.Statements[0]
		if !testInsertStatement(t, stmt, tt.expectedIdentifier, tt.expectedVal) {
			return
		}
	}
}

func testInsertStatement(t *testing.T, s ast.Statement, name string, val []string) bool {
	if s.TokenLiteral() != "INSERT" {
		t.Errorf("s.TokenLiteral not 'let'. got=%q", s.TokenLiteral())
		return false
	}

	letStmt, ok := s.(*ast.InsertStatement)
	if !ok {
		t.Errorf("s not *ast.Delete. got=%T", s)
		return false
	}

	if letStmt.TName.Val != name {
		t.Errorf("letStmt.Name.Value not '%s'. got=%s", name, letStmt.TName.Val)
		return false
	}

	for i := range val {
		if letStmt.Values[i] != val[i] {
			t.Errorf("Where clause CName value expected: '%s'. got=%s", val[i], letStmt.Values[i])
			return false
		}
	}

	return true
}

func TestDeleteStatement(t *testing.T) {
	tests := []struct {
		input              string
		expectedIdentifier string
		expectedCol        []string
		expectedVal        []string
	}{
		{"DELETE FROM dogs WHERE name = 'stella';", "dogs", []string{"name"}, []string{"stella"}},
		{"DELETE FROM dogs WHERE name = 'stella' AND breed = 'labradoodle';", "dogs", []string{"name", "breed"}, []string{"stella", "labradoodle"}},
		{"DELETE FROM dogs WHERE name = 'stella' AND breed = 'labradoodle' AND age = 7;", "dogs", []string{"name", "breed", "age"}, []string{"stella", "labradoodle", "7"}},
	}

	for _, tt := range tests {
		program := createParseProgram(tt.input, t)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain 1 statements. got=%d", len(program.Statements))
		}

		stmt := program.Statements[0]
		if !testDeleteStatement(t, stmt, tt.expectedIdentifier, tt.expectedCol, tt.expectedVal) {
			return
		}
	}
}

func testDeleteStatement(t *testing.T, s ast.Statement, name string, cols, val []string) bool {
	if s.TokenLiteral() != "DELETE" {
		t.Errorf("s.TokenLiteral not 'let'. got=%q", s.TokenLiteral())
		return false
	}

	letStmt, ok := s.(*ast.DeleteStatement)
	if !ok {
		t.Errorf("s not *ast.Delete. got=%T", s)
		return false
	}

	if letStmt.TName.Val != name {
		t.Errorf("letStmt.Name.Value not '%s'. got=%s", name, letStmt.TName.Val)
		return false
	}

	if len(cols) > 1 || len(val) > 1 {
		if len(cols) != len(val) {
			t.Errorf("Expected cols and val to have same length, but cols length was: %d while vals length was: %d", len(cols), len(val))
			return false
		}
		for i := range cols {
			if letStmt.Condition[i].CName.Val != cols[i] {
				t.Errorf("Where clause CName value expected: '%s'. got=%s", cols[i], letStmt.Condition[i].CName.Val)
				return false
			}

			if letStmt.Condition[i].CIdent.Val != val[i] {
				t.Errorf("Where clause identifier value expected: '%s'. got=%s", val, letStmt.Condition[i].CIdent.Val)
				return false
			}
		}
	} else {
		if letStmt.Condition[0].CName.Val != cols[0] {
			t.Errorf("Where clause CName value expected: '%s'. got=%s", cols[0], letStmt.Condition[0].CIdent.Val)
			return false
		}

		if letStmt.Condition[0].CIdent.Val != val[0] {
			t.Errorf("Where clause CIdent value expected: '%s'. got=%s", val, letStmt.Condition[0].CName.Val)
			return false
		}
	}

	return true
}

func TestSelectStatement(t *testing.T) {
	tests := []struct {
		input              string
		expectedCols       string
		expectedIdentifier string
		expectedCol        []string
		expectedVal        []string
	}{
		{"SELECT * FROM dogs WHERE name = 'stella';", "*", "dogs", []string{"name"}, []string{"stella"}},
		{"SELECT * FROM dogs WHERE name = 'stella' AND breed = 'labradoodle';", "*", "dogs", []string{"name", "breed"}, []string{"stella", "labradoodle"}},
		{"SELECT * FROM dogs WHERE name = 'stella' AND breed = 'labradoodle' AND age = 7;", "*", "dogs", []string{"name", "breed", "age"}, []string{"stella", "labradoodle", "7"}},
	}

	for _, tt := range tests {
		program := createParseProgram(tt.input, t)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain 1 statements. got=%d", len(program.Statements))
		}

		stmt := program.Statements[0]
		if !testSelectStatement(t, stmt, tt.expectedIdentifier, tt.expectedCol, tt.expectedVal) {
			return
		}
	}
}

func testSelectStatement(t *testing.T, s ast.Statement, name string, cols, val []string) bool {
	if s.TokenLiteral() != "SELECT" {
		t.Errorf("s.TokenLiteral not SELECt. got=%q", s.TokenLiteral())
		return false
	}

	letStmt, ok := s.(*ast.SelectStatement)
	if !ok {
		t.Errorf("s not *ast.LetStatement. got=%T", s)
		return false
	}
	// if letStmt.NumCols.Literal != "*" {
	// 	t.Errorf("letStmt.Name.Value not '%s'. got=%s", cols, letStmt.NumCols)
	// 	return false
	// }

	// if letStmt.From.Literal != "FROM" {
	// 	t.Errorf("letStmt.Name.Value not '%s'. got=%s", "FROM", letStmt.From.Literal)
	// 	return false
	// }

	if letStmt.TName.Val != name {
		t.Errorf("letStmt.Name.Value not '%s'. got=%s", name, letStmt.TName.Val)
		return false
	}
	// if letStmt.Where.Literal != "WHERE" {
	// 	t.Errorf("Where clause expected 'WHERE' token. got=%s", letStmt.TName.Val)
	// 	return false
	// }

	if len(cols) > 1 || len(val) > 1 {
		if len(cols) != len(val) {
			t.Errorf("Expected cols and val to have same length, but cols length was: %d while vals length was: %d", len(cols), len(val))
			return false
		}
		for i := range cols {
			if letStmt.Condition[i].CName.Val != cols[i] {
				t.Errorf("Where clause CName value expected: '%s'. got=%s", cols[i], letStmt.Condition[i].CName.Val)
				return false
			}

			if letStmt.Condition[i].CIdent.Val != val[i] {
				t.Errorf("Where clause identifier value expected: '%s'. got=%s", val, letStmt.Condition[i].CIdent.Val)
				return false
			}
		}
	} else {
		if letStmt.Condition[0].CName.Val != cols[0] {
			t.Errorf("Where clause CName value expected: '%s'. got=%s", cols[0], letStmt.Condition[0].CIdent.Val)
			return false
		}

		if letStmt.Condition[0].CIdent.Val != val[0] {
			t.Errorf("Where clause CIdent value expected: '%s'. got=%s", val, letStmt.Condition[0].CName.Val)
			return false
		}
	}

	return true
}

func TestCreateTableStatement(t *testing.T) {
	tests := []struct {
		input             string
		expectedTableName string
		expectedCols      []string
		expectedTypes     []string
	}{
		{"CREATE TABLE dogs (name varchar, breed varchar);", "dogs", []string{"name", "breed"}, []string{"varchar", "varchar"}},
	}

	for _, tt := range tests {
		program := createParseProgram(tt.input, t)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain 1 statements. got=%d", len(program.Statements))
		}

		stmt := program.Statements[0]
		if !testTableStatement(t, stmt, tt.expectedTableName, tt.expectedCols, tt.expectedTypes) {
			return
		}
	}
}

func testTableStatement(t *testing.T, s ast.Statement, name string, cols, types []string) bool {
	if s.TokenLiteral() != "CREATE" {
		t.Errorf("s.TokenLiteral not 'let'. got=%q", s.TokenLiteral())
		return false
	}

	letStmt, ok := s.(*ast.CreateTableStatement)
	if !ok {
		t.Errorf("s not *ast.Delete. got=%T", s)
		return false
	}

	if letStmt.TName.Val != name {
		t.Errorf("letStmt.Name.Value not '%s'. got=%s", name, letStmt.TName.Val)
		return false
	}

	for i := range cols {
		if letStmt.Cols[i] != cols[i] {
			t.Errorf("Where clause CName value expected: '%s'. got=%s", cols[i], letStmt.Cols[i])
			return false
		}
		if letStmt.ColTypes[i] != types[i] {
			t.Errorf("Where clause CName value expected: '%s'. got=%s", types[i], letStmt.ColTypes[i])
			return false
		}
	}
	return true
}

func TestCreateIndexStatement(t *testing.T) {
	tests := []struct {
		input             string
		expectedTableName string
		expectedCols      []string
	}{
		{"CREATE INDEX ON dogs (name);", "dogs", []string{"name"}},
		{"CREATE INDEX ON dogs (name, breed);", "dogs", []string{"name", "breed"}},
	}

	for _, tt := range tests {
		program := createParseProgram(tt.input, t)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain 1 statements. got=%d", len(program.Statements))
		}

		stmt := program.Statements[0]
		if !testCreateIndexStatement(t, stmt, tt.expectedTableName, tt.expectedCols) {
			return
		}
	}
}

func testCreateIndexStatement(t *testing.T, s ast.Statement, name string, cols []string) bool {
	if s.TokenLiteral() != "CREATE" {
		t.Errorf("s.TokenLiteral not 'let'. got=%q", s.TokenLiteral())
		return false
	}

	letStmt, ok := s.(*ast.CreateIndexStatement)
	if !ok {
		t.Errorf("s not *ast.Delete. got=%T", s)
		return false
	}

	if letStmt.TName.Val != name {
		t.Errorf("letStmt.Name.Value not '%s'. got=%s", name, letStmt.TName.Val)
		return false
	}

	for i := range cols {
		if letStmt.Cols[i].Val != cols[i] {
			t.Errorf("Where clause CName value expected: '%s'. got=%s", cols[i], letStmt.Cols[i])
			return false
		}
	}
	return true
}

func checkParserErrors(t *testing.T, p *Parser) {
	errors := p.Errs()
	if len(errors) == 0 {
		return
	}

	t.Errorf("parser has %d erros", len(errors))
	for _, msg := range errors {
		t.Errorf("parser error: %q", msg)
	}
	t.FailNow()
}
