package compile

import (
	"testing"

	"github.com/aidanjjenkins/compiler/ast"
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
		input       string
		expectedVal []string
	}{
		{"CREATE TABLE dogs (name varchar, breed varchar);", []string{"dogs", "name", "varchar", "false", "false", "false", "breed", "varchar", "false", "false", "false"}},
	}

	for _, tt := range tests {
		program := createParseProgram(tt.input, t)
		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain 1 statements. got=%d", len(program.Statements))
		}

		stmt := program.Statements[0]
		comp := New()

		if !testAddTableStatement(t, stmt, comp, tt.expectedVal) {
			return
		}

	}
}

func testAddTableStatement(t *testing.T, stmt ast.Statement, comp *Compiler, val []string) bool {
	comp.Compile(stmt)

	op1 := comp.Instructions[0]
	if op1 != 6 {
		t.Errorf("Expected: 6, got: %d", op1)
		return false
	}
	op2 := comp.Instructions[1]
	if op2 != 12 {
		t.Errorf("Expected: 12, got: %d", op2)
		return false
	}

	decoded := DecodeBytes(comp.Writes[0][TotalLengths:])
	for i := range decoded {
		if decoded[i] != val[i] {
			t.Errorf("Expected: %s, got: %s", val[i], decoded[i])
			return false
		}
	}
	return true
}
