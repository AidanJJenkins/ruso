package vm

import (
	"os"
	"testing"

	"github.com/aidanjjenkins/compiler/ast"
	c "github.com/aidanjjenkins/compiler/compile"
	"github.com/aidanjjenkins/compiler/lexer"
	"github.com/aidanjjenkins/compiler/parser"
)

func createParseProgram(input string, t *testing.T) *ast.Program {
	l := lexer.New(input)
	p := parser.New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	return program
}

func checkParserErrors(t *testing.T, p *parser.Parser) {
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

func TestAddTable(t *testing.T) {
	tests := []struct {
		input string
		name  string
	}{
		{"CREATE TABLE dogs (name varchar, breed varchar);", "dogs"},
		// {"CREATE TABLE people (age varchar, height varchar);"},
	}

	for _, tt := range tests {
		program := createParseProgram(tt.input, t)
		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain 1 statements. got=%d", len(program.Statements))
		}

		stmt := program.Statements[0]
		comp := c.New()

		if !testAddTableStatement(t, stmt, comp, tt.name) {
			return
		}

	}
}

func testAddTableStatement(t *testing.T, stmt ast.Statement, comp *c.Compiler, n string) bool {
	err := comp.Compile(stmt)
	if err != nil {
		t.Error("Compile error: ", err)
		return false
	}

	machine := New(comp.Bytecode())
	err = machine.Run()
	if err != nil {
		t.Error("Error running")
		return false
	}

	row := machine.FindTable("dogs")
	if row == nil {
		t.Errorf("Row not found")
		return false
	}

	name := GetTableNameBytes(row)
	if name != n {
		t.Errorf("Expected name to be: %s, got: %s", n, name)
		return false
	}

	os.Remove("db.db")
	os.Remove("test.db")

	return true
}
