package compile

import (
	// "fmt"
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
	}{
		{"CREATE TABLE dogs (name varchar, breed varchar);", "dogs", []string{"name", "varchar", "breed", "varchar"}},
		{"CREATE TABLE dogs (name varchar, breed varchar, age varchar);", "dogs", []string{"name", "varchar", "breed", "varchar", "age", "varchar"}},
		{"CREATE TABLE students (grade varchar, subject varchar);", "students", []string{"grade", "varchar", "subject", "varchar"}},
	}

	for _, tt := range tests {
		program := createParseProgram(tt.input, t)
		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain 1 statements. got=%d", len(program.Statements))
		}

		stmt := program.Statements[0]
		comp := New()

		if !testAddTableStatement(t, stmt, comp, tt.expectedTable, tt.expectedVal) {
			return
		}

	}
}

func testAddTableStatement(t *testing.T, stmt ast.Statement, comp *Compiler, name string, val []string) bool {
	comp.Compile(stmt)

	n := AccessTableNameBytes(comp.Instructions)
	if name != n[:len(name)] {
		t.Errorf("Expected: %s, got: %s", name, n)
		return false
	}

	firstByte := AccessFirstByte(comp.Instructions)
	if firstByte != 6 {
		t.Errorf("Expected: 6, got: %d", firstByte)
		return false
	}
	// AccessTableMetaDataIndex(comp.Instructions)
	lens := AccessTableMetaDataLengths(comp.Instructions)
	vals := AccessTableRowInfoBytes(comp.Instructions[TableMetaDataSize+256:], lens)

	for i := range val {
		if val[i] != vals[i] {
			t.Errorf("Expected: %s, got: %s", val[i], vals[i])
			return false
		}
	}

	return true
}

func TestCreateIndex(t *testing.T) {
	tests := []struct {
		input         string
		expectedTable string
		expectedVal   []string
	}{
		{"CREATE INDEX ON dogs (name);", "dogs", []string{"name"}},
		{"CREATE INDEX ON dogs (name, breed, age);", "dogs", []string{"name", "breed", "age"}},
	}

	for _, tt := range tests {
		program := createParseProgram(tt.input, t)
		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain 1 statements. got=%d", len(program.Statements))
		}

		stmt := program.Statements[0]
		comp := New()

		if !testCreateIndex(t, stmt, comp, tt.expectedTable, tt.expectedVal) {
			return
		}

	}
}

func testCreateIndex(t *testing.T, stmt ast.Statement, comp *Compiler, name string, val []string) bool {
	comp.Compile(stmt)

	ins := AccessFirstByte(comp.Instructions)
	if ins != 7 {
		t.Errorf("Expected: 6, got: %d", ins)
		return false
	}

	tablename := AccessIndexName(comp.Instructions)
	if name != tablename[:len(name)] {
		t.Errorf("Expected: %s, got: %s", name, tablename)
		return false
	}

	vals := AccessIndexVals(comp.Instructions)

	for i := range val {
		n := vals[i]
		if val[i] != n[:len(val[i])] {
			t.Errorf("Expected: %s, got: %s", val[i], n)
			return false
		}
	}

	return true
}

func TestSelect(t *testing.T) {
	tests := []struct {
		input         string
		expectedTable string
		expectedVal   []string
	}{
		{"SELECT * FROM dogs WHERE name = \"stella\";", "dogs", []string{"name", "stella"}},
		{"SELECT * FROM dogs WHERE name = \"winnie\" AND breed = \"cane corso\";", "dogs", []string{"name", "winnie", "breed", "cane corso"}},
		{"SELECT * FROM dogs WHERE age = \"3\" AND breed = \"cane corso\";", "dogs", []string{"age", "3", "breed", "cane corso"}},
	}

	for _, tt := range tests {
		program := createParseProgram(tt.input, t)
		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain 1 statements. got=%d", len(program.Statements))
		}

		stmt := program.Statements[0]
		comp := New()

		if !testSelect(t, stmt, comp, tt.expectedTable, tt.expectedVal) {
			return
		}

	}
}

func testSelect(t *testing.T, stmt ast.Statement, comp *Compiler, name string, colval []string) bool {
	comp.Compile(stmt)

	ins := AccessFirstByte(comp.Instructions)
	if ins != 8 {
		t.Errorf("Expected: 6, got: %d", ins)
		return false
	}

	vals := AccessSelectValues(comp.Instructions)
	n := vals[0]
	if name != n[:len(name)] {
		t.Errorf("Expected: %s, got: %s", name, n)
		return false
	}
	vals = vals[1:]

	for i := range colval {
		inputVal := vals[i]
		if colval[i] != inputVal[:len(colval[i])] {
			t.Errorf("Expected: %s, got: %s", colval[i], inputVal)
			return false
		}
	}

	return true
}

func TestInsert(t *testing.T) {
	tests := []struct {
		input         string
		expectedTable string
		expectedVal   []string
	}{
		{"INSERT INTO dogs (\"stella\", \"labradoodle\");", "dogs", []string{"stella", "labradoodle"}},
		{"INSERT INTO dogs (\"winnie\", \"cane corso\", \"3\" );", "dogs", []string{"winnie", "cane corso", "3"}},
	}

	for _, tt := range tests {
		program := createParseProgram(tt.input, t)
		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain 1 statements. got=%d", len(program.Statements))
		}

		stmt := program.Statements[0]
		comp := New()

		if !testInsert(t, stmt, comp, tt.expectedTable, tt.expectedVal) {
			return
		}

	}
}

func testInsert(t *testing.T, stmt ast.Statement, comp *Compiler, name string, vals []string) bool {
	comp.Compile(stmt)

	ins := AccessFirstByte(comp.Instructions)
	if ins != 9 {
		t.Errorf("Expected: 6, got: %d", ins)
		return false
	}

	iv := AccessSelectValues(comp.Instructions)
	n := iv[0]
	if name != n[:len(name)] {
		t.Errorf("Expected: %s, got: %s", name, n)
		return false
	}
	iv = iv[1:]

	for i := range vals {
		inputVal := vals[i]
		if vals[i] != inputVal[:len(vals[i])] {
			t.Errorf("Expected: %s, got: %s", vals[i], inputVal)
			return false
		}
	}

	return true
}
