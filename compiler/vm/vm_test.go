package vm

import (
	"fmt"
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
		{"CREATE TABLE wishlist (brand varchar, price varchar, seller varchar);", "wishlist"},
		{"CREATE TABLE people (brand varchar, price varchar, seller varchar);", "people"},
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
	os.Remove(TableFile)
	os.Remove(IdxFile)
}

func testAddTableStatement(t *testing.T, stmt ast.Statement, comp *c.Compiler, n string) bool {
	err := comp.Compile(stmt)
	if err != nil {
		t.Error("Compile error: ", err)
		return false
	}

	tInfo := make(map[string]*Tbls)

	machine := New(comp.Bytecode(), tInfo)
	err = machine.Run()
	if err != nil {
		t.Error("Error running")
		return false
	}

	row := machine.FindTable(n)
	if row == nil {
		t.Errorf("Row not found")
		return false
	}

	name := GetTableNameBytes(row)
	if name != n {
		t.Errorf("Expected name to be: %s, got: %s", n, name)
		return false
	}

	return true
}

func dummy(t *testing.T) {
	tests := []struct {
		input string
	}{
		{"CREATE TABLE wishlist (name varchar);"},
		{"CREATE TABLE people (age varchar, name varchar);"},
	}

	for _, tt := range tests {
		program := createParseProgram(tt.input, t)
		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain 1 statements. got=%d", len(program.Statements))
		}

		stmt := program.Statements[0]
		comp := c.New()

		err := comp.Compile(stmt)
		if err != nil {
			t.Error("Compile error: ", err)
			return
		}

		tInfo := make(map[string]*Tbls)
		machine := New(comp.Bytecode(), tInfo)
		err = machine.Run()
		if err != nil {
			t.Error("Error running")
			return
		}
	}
}

func TestAddIndex(t *testing.T) {
	dummy(t)
	err := printTableFile()
	if err != nil {
		fmt.Println("error printing table file: ", err)
	}
	tests := []struct {
		input string
	}{
		// {"CREATE INDEX ON wishlist (name);"},
		{"CREATE INDEX ON people (age);"},
	}

	for _, tt := range tests {
		program := createParseProgram(tt.input, t)
		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain 1 statements. got=%d", len(program.Statements))
		}

		stmt := program.Statements[0]
		comp := c.New()

		err := comp.Compile(stmt)
		if err != nil {
			t.Error("Compile error: ", err)
			return
		}

		tInfo := make(map[string]*Tbls)
		machine := New(comp.Bytecode(), tInfo)
		err = machine.Run()
		if err != nil {
			t.Error("Error running")
			return
		}
	}
	err = printTableFile()
	if err != nil {
		fmt.Println("error printing table file: ", err)
	}
	os.Remove(TableFile)
	os.Remove(IdxFile)
}

func printTableFile() error {
	file, err := os.Open(TableFile)
	if err != nil {
		return err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Println("Next offset is not withing file size", err)
		return err
	}

	fileSize := fileInfo.Size()
	fileBytes := make([]byte, fileSize)
	file.Read(fileBytes)
	// fmt.Println("file: ", fileBytes)
	return nil
}

// func TestSelect(t *testing.T) {
// 	tests := []struct {
// 		input string
// 		name  string
// 		where []string
// 	}{
// 		{"SELECT * FROM dogs WHERE name = \"stella\";", "dogs", []string{"name", "stella"}},
// 	}
//
// 	for _, tt := range tests {
// 		program := createParseProgram(tt.input, t)
// 		if len(program.Statements) != 1 {
// 			t.Fatalf("program.Statements does not contain 1 statements. got=%d", len(program.Statements))
// 		}
//
// 		stmt := program.Statements[0]
// 		comp := c.New()
//
// 		if !testSelectStatement(t, stmt, comp, tt.name, tt.where) {
// 			return
// 		}
//
// 	}
// }
//
// func testSelectStatement(t *testing.T, stmt ast.Statement, comp *c.Compiler, n string, w []string) bool {
// 	err := comp.Compile(stmt)
// 	if err != nil {
// 		t.Error("Compile error: ", err)
// 		return false
// 	}
//
// 	machine := New(comp.Bytecode())
// 	err = machine.Run()
// 	if err != nil {
// 		t.Error("Error running")
// 		return false
// 	}
//
// 	os.Remove("db.db")
// 	os.Remove("test.db")
//
// 	return true
// }
