package compile

import (
	"fmt"
	"testing"

	"github.com/aidanjjenkins/compiler/ast"
	"github.com/aidanjjenkins/compiler/code"
	"github.com/aidanjjenkins/compiler/lexer"
	"github.com/aidanjjenkins/compiler/parser"
)

func TestIndex(t *testing.T) {
	name := code.TableName{Value: "dogs"}
	col := code.Col{Value: "name"}
	tests := []compilerTestCase{
		{
			input:             "CREATE INDEX ON dogs (name);",
			expectedConstants: []interface{}{col, name},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpAddIndex, 0),
				code.Make(code.OpCreateTableIndex, 1),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestCreateTable(t *testing.T) {
	cell := code.ColCell{Name: "name", ColType: "varchar", Unique: false, Index: false, Pk: false}
	name := code.TableName{Value: "people"}
	tests := []compilerTestCase{
		{
			input:             "CREATE TABLE people (name varchar);",
			expectedConstants: []interface{}{name, cell},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpEncodeStringVal, 0),
				code.Make(code.OpEncodeTableCell, 1),
				code.Make(code.OpCreateTable, 2),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestInsertRow(t *testing.T) {
	name := code.TableName{Value: "dogs"}
	col1 := code.Col{Value: "stella"}
	col2 := code.Col{Value: "labradoodle"}
	tests := []compilerTestCase{
		{
			input:             "INSERT INTO dogs (\"stella\", \"labradoodle\");",
			expectedConstants: []interface{}{name, col1, col2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpEncodeStringVal, 0),
				code.Make(code.OpEncodeStringVal, 1),
				code.Make(code.OpEncodeStringVal, 2),
				code.Make(code.OpInsertRow, 3),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestSelect(t *testing.T) {
	where := code.Where{Column: "name", Value: "rtx 4090"}
	name := code.TableName{Value: "wishlist"}
	tests := []compilerTestCase{
		{
			input:             "SELECT * FROM wishlist WHERE name = \"rtx 4090\";",
			expectedConstants: []interface{}{name, where},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpWhereCondition, 1),
				code.Make(code.OpSelect, 2),
			},
		},
	}

	runCompilerTests(t, tests)
}

type compilerTestCase struct {
	input                string
	expectedConstants    []interface{}
	expectedInstructions []code.Instructions
}

func parse(input string) *ast.Program {
	l := lexer.New(input)
	p := parser.New(l)
	return p.ParseProgram()
}

func runCompilerTests(t *testing.T, tests []compilerTestCase) {
	t.Helper()

	for _, tt := range tests {
		program := parse(tt.input)

		compiler := New()
		err := compiler.Compile(program)
		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}

		bytecode := compiler.Bytecode()

		err = testInstructions(tt.expectedInstructions, bytecode.Instructions)
		if err != nil {
			t.Fatalf("testInstructions failed: %s", err)
		}

		err = testConstants(t, tt.expectedConstants, bytecode.constants)
		if err != nil {
			t.Fatalf("testConstants failed: %s", err)
		}
	}
}

func testInstructions(
	expected []code.Instructions,
	actual code.Instructions,
) error {
	concatted := concatInstructions(expected)

	if len(actual) != len(concatted) {
		return fmt.Errorf("wrong instructions length.\nwant=%q\ngot =%q",
			concatted, actual)
	}

	for i, ins := range concatted {
		if actual[i] != ins {
			return fmt.Errorf("wrong instruction at %d.\nwant=%q\ngot =%q",
				i, concatted, actual)
		}
	}

	return nil
}

func concatInstructions(s []code.Instructions) code.Instructions {
	out := code.Instructions{}

	for _, ins := range s {
		out = append(out, ins...)
	}

	return out
}

func testConstants(
	t *testing.T,
	expected []interface{},
	actual []code.Obj,
) error {
	if len(expected) != len(actual) {
		return fmt.Errorf("wrong number of constants. got=%d, want=%d",
			len(actual), len(expected))
	}

	for i, constant := range expected {
		switch constant := constant.(type) {
		case code.TableName:
			err := testTableName(constant, actual[i])
			if err != nil {
				return fmt.Errorf("constant %d - test Index failed: %s",
					i, err)
			}
		case code.ColCell:
			err := testColCell(constant, actual[i])
			if err != nil {
				return fmt.Errorf("constant %d - test Index failed: %s",
					i, err)
			}
		case code.Col:
			err := testCol(constant, actual[i])
			if err != nil {
				return fmt.Errorf("constant %d - test Index failed: %s",
					i, err)
			}
		case code.Where:
			err := testWhere(constant, actual[i])
			if err != nil {
				return fmt.Errorf("constant %d - test Index failed: %s",
					i, err)
			}
		}
	}

	return nil
}

func testTableName(expected code.TableName, actual code.Obj) error {
	result, ok := actual.(*code.TableName)
	if !ok {
		return fmt.Errorf("object is not same type. got=%T (%+v)",
			actual, actual)
	}

	if result.Value != expected.Value {
		return fmt.Errorf("Expected: %s, got: %s",
			result.Value, result.Value)
	}

	return nil
}

func testColCell(expected code.ColCell, actual code.Obj) error {
	result, ok := actual.(*code.ColCell)
	if !ok {
		return fmt.Errorf("object is not same type. got=%T (%+v)",
			actual, actual)
	}

	if result.Name != expected.Name {
		return fmt.Errorf("Expected: %s, got: %s",
			result.Name, result.Name)
	}

	if result.ColType != expected.ColType {
		return fmt.Errorf("Expected: %s, got: %s",
			result.ColType, result.ColType)
	}

	return nil
}

func testCol(expected code.Col, actual code.Obj) error {
	result, ok := actual.(*code.Col)
	if !ok {
		return fmt.Errorf("object is not same type. got=%T (%+v)",
			actual, actual)
	}

	if result.Value != expected.Value {
		return fmt.Errorf("Expected: %s, got: %s",
			result.Value, result.Value)
	}
	return nil
}

func testWhere(expected code.Where, actual code.Obj) error {
	result, ok := actual.(*code.Where)
	if !ok {
		return fmt.Errorf("object is not same type. got=%T (%+v)",
			actual, actual)
	}

	if result.Column != expected.Column {
		return fmt.Errorf("Expected: %s, got: %s",
			expected.Column, result.Column)
	}

	if result.Value != expected.Value {
		return fmt.Errorf("Expected: %s, got: %s",
			expected.Value, result.Value)
	}

	return nil
}

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
