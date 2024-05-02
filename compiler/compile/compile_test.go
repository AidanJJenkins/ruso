package compile

import (
	"fmt"
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
		input         string
		expectedTable string
		expectedVal   []string
		expectedType  []string
	}{
		{"CREATE TABLE dogs (name varchar, breed varchar);", "dogs", []string{"name", "breed"}, []string{"varchar", "varchar"}},
	}

	program := createParseProgram(tests[0].input, t)
	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statements. got=%d", len(program.Statements))
	}

	stmt := program.Statements[0]
	comp := New()

	// if !testInsertStatement(t, stmt, comp, tt.expectedTable, tt.expectedVal, tt.expectedType) {
	// 	return
	// }
	testInsertStatement(t, stmt, comp, tests[0].expectedTable, tests[0].expectedVal, tests[0].expectedType)
}

func testInsertStatement(t *testing.T, stmt ast.Statement, comp *Compiler, name string, val, types []string) bool {
	fmt.Println("name", name)
	fmt.Println("val", val)
	fmt.Println("type", types)

	comp.Compile(stmt)

	// fmt.Println("opcode: ", comp.Instructions[0])
	// fmt.Println("opcode: ", string(comp.Instructions[1:]))
	return false
}
