package compile

import (
	"fmt"
	"testing"

	"github.com/aidanjjenkins/compiler/ast"
	"github.com/aidanjjenkins/compiler/code"
	"github.com/aidanjjenkins/compiler/lexer"
	o "github.com/aidanjjenkins/compiler/object"
	"github.com/aidanjjenkins/compiler/parser"
)

func TestIndex(t *testing.T) {
	name := o.TableName{Value: "dogs"}
	col := o.Col{Value: "name"}
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
	cell := o.ColCell{Name: "name", ColType: "varchar", Unique: false, Index: false, Pk: false}
	name := o.TableName{Value: "people"}
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
	name := o.TableName{Value: "dogs"}
	col1 := o.Col{Value: "stella"}
	col2 := o.Col{Value: "labradoodle"}
	tests := []compilerTestCase{
		{
			input:             "INSERT INTO dogs VALUES (\"stella\", \"labradoodle\");",
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

func TestInsertRowDouble(t *testing.T) {
	name := o.TableName{Value: "dogs"}
	col1 := o.Col{Value: "col1"}
	col2 := o.Col{Value: "col2"}
	val1 := o.Col{Value: "val1"}
	val2 := o.Col{Value: "val2"}
	tests := []compilerTestCase{
		{
			input:             "INSERT INTO dogs (col1, col2) VALUES (\"val1\", \"val2\");",
			expectedConstants: []interface{}{name, col1, col2, val1, val2},
			expectedInstructions: []code.Instructions{

				// create table object, get table and adds its columns to an array, leave rest null, push to stack
				// as you get check col instructions, mark an array with positions corresponding to what order was given
				// if a table has cols: name, age, breed, weight
				// and the given cols are: age, name, weight
				// mark an array as [0, 0, 0, 0]
				//[0, 1, 0, 0]
				//[2, 1, 0, 0]
				//[2, 1, 0, 3]
				// as you get the values to write, add them to an array that gets pushed to the stack,
				// they should be encoded to bytes, then added to an an array of an array of bytes
				//[7]
				//[7, stella]
				//[7, stella, 20]
				// when done, loop through the values, add 1 to the loop's current index,
				// and see where that number pops up in the marked array, thats it location in the final
				//write

				// as you go, the table info can have fields for the column traits,
				// for example: "index" : [false, true, false, false]
				// before you write, you can check the index field, and go to the second item, and insert it into the index

				code.Make(code.OpEncodeStringVal, 0),
				code.Make(code.OpCheckCol, 1),
				code.Make(code.OpCheckCol, 2),
				code.Make(code.OpEncodeStringVal, 3),
				code.Make(code.OpEncodeStringVal, 4),
				code.Make(code.OpInsertRow, 5),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestSelect(t *testing.T) {
	where := o.Where{Column: "name", Value: "rtx 4090"}
	name := o.TableName{Value: "wishlist"}
	tests := []compilerTestCase{
		{
			input:             "SELECT * FROM wishlist WHERE name = \"rtx 4090\";",
			expectedConstants: []interface{}{name, where},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpTableNameSearch, 0),
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

		err = testConstants(t, tt.expectedConstants, bytecode.Constants)
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
	// fmt.Println("concatted expected: ", concatted)
	// fmt.Println("actual: ", actual)
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
	actual []o.Obj,
) error {
	if len(expected) != len(actual) {
		return fmt.Errorf("wrong number of constants. got=%d, want=%d",
			len(actual), len(expected))
	}

	for i, constant := range expected {
		switch constant := constant.(type) {
		case o.TableName:
			err := testTableName(constant, actual[i])
			if err != nil {
				return fmt.Errorf("constant %d - test Index failed: %s",
					i, err)
			}
		case o.ColCell:
			err := testColCell(constant, actual[i])
			if err != nil {
				return fmt.Errorf("constant %d - test Index failed: %s",
					i, err)
			}
		case o.Col:
			err := testCol(constant, actual[i])
			if err != nil {
				return fmt.Errorf("constant %d - test Index failed: %s",
					i, err)
			}
		case o.Where:
			err := testWhere(constant, actual[i])
			if err != nil {
				return fmt.Errorf("constant %d - test Index failed: %s",
					i, err)
			}
		}
	}

	return nil
}

func testTableName(expected o.TableName, actual o.Obj) error {
	result, ok := actual.(*o.TableName)
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

func testColCell(expected o.ColCell, actual o.Obj) error {
	result, ok := actual.(*o.ColCell)
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

func testCol(expected o.Col, actual o.Obj) error {
	result, ok := actual.(*o.Col)
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

func testWhere(expected o.Where, actual o.Obj) error {
	result, ok := actual.(*o.Where)
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
