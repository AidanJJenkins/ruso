package vm

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/aidanjjenkins/compiler/ast"
	"github.com/aidanjjenkins/compiler/code"
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
		input        string
		expectedVals []string
	}{
		{"CREATE TABLE dogs (name varchar, breed varchar);", []string{"dogs", "name", "varchar", "false", "false", "false", "breed", "varchar", "false", "false", "false"}},
	}

	for _, tt := range tests {
		program := createParseProgram(tt.input, t)
		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain 1 statements. got=%d", len(program.Statements))
		}

		stmt := program.Statements[0]
		comp := c.New()

		if !testAddTableStatement(t, stmt, comp, tt.expectedVals) {
			return
		}

	}

	os.Remove(IdxFile)
	os.Remove(TableFile)
}

func testAddTableStatement(t *testing.T, stmt ast.Statement, comp *c.Compiler, vals []string) bool {
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

	offset, ok := machine.Pool.Search("dogs")
	decoded := []string{}
	var count int
	if !ok {
		fmt.Println("Table not found")
	} else {
		row, err := readRow(int64(offset), TableFile)
		if err != nil {
			fmt.Println("Error finding table: ", err)
		}

		decoded = DecodeBytes(row)
		count = DecodeTableCount(row)
	}

	if count != 0 {
		t.Errorf("Expected: 0, got: %d", count)
		return false
	}
	for i := range vals {
		if vals[i] != decoded[i] {
			t.Errorf("Expected: %s, got: %s", vals[i], decoded[i])
			return false
		}
	}
	return true
}

func TestInsert(t *testing.T) {
	createDummyTablesForInsert(t)
	tests := []struct {
		input        string
		expectedVals []string
		count        int
	}{
		{"INSERT INTO wishlist VALUES (\"4090\", \"1000\");", []string{"wishlist", "4090", "1000", " "}, 1},
	}

	for _, tt := range tests {
		program := createParseProgram(tt.input, t)
		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain 1 statements. got=%d", len(program.Statements))
		}

		stmt := program.Statements[0]
		comp := c.New()

		if !testInsert(t, stmt, comp, tt.expectedVals, tt.count) {
			os.Remove(IdxFile)
			os.Remove(RowsFile)
			os.Remove(TableFile)
			return
		}

	}

	os.Remove(IdxFile)
	os.Remove(RowsFile)
	os.Remove(TableFile)
}

func testInsert(t *testing.T, stmt ast.Statement, comp *c.Compiler, expected []string, count int) bool {
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

	lastRow, err := getLastRow(RowsFile)
	if err != nil {
		fmt.Printf("Error getting last row: %v\n", err)
		return false
	}

	for i := range expected {
		if expected[i] != lastRow[i] {
			t.Errorf("Expected: %s, got: %s", expected[i], lastRow[i])
			return false
		}
	}

	offset, ok := machine.Pool.Search(expected[0])
	var c int
	if !ok {
		fmt.Println("Table not found")
	} else {
		row, err := readRow(int64(offset), TableFile)
		if err != nil {
			fmt.Println("Error finding table: ", err)
		}

		c = DecodeTableCount(row)
		if c != count {
			t.Errorf("Expected: %d, got: %d", count, c)
			return false
		}
	}

	return true
}

func TestInsertDouble(t *testing.T) {
	createDummyTablesForInsert(t)
	tests := []struct {
		input        string
		expectedVals []string
		count        int
	}{
		{"INSERT INTO wishlist (name, price) VALUES (\"4090\", \"1000\");", []string{"wishlist", "4090", " ", "1000"}, 1},
		{"INSERT INTO wishlist (name, brand, price) VALUES (\"4090\", \"nvidia\",\"1000\");", []string{"wishlist", "4090", "nvidia", "1000"}, 2},
	}

	for _, tt := range tests {
		program := createParseProgram(tt.input, t)
		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain 1 statements. got=%d", len(program.Statements))
		}

		stmt := program.Statements[0]
		comp := c.New()

		if !testInsertDouble(t, stmt, comp, tt.expectedVals, tt.count) {
			os.Remove(IdxFile)
			os.Remove(RowsFile)
			os.Remove(TableFile)
			return
		}

	}

	os.Remove(IdxFile)
	os.Remove(RowsFile)
	os.Remove(TableFile)
}

func testInsertDouble(t *testing.T, stmt ast.Statement, comp *c.Compiler, expected []string, count int) bool {
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

	lastRow, err := getLastRow(RowsFile)
	if err != nil {
		fmt.Printf("Error getting last row: %v\n", err)
		return false
	}

	for i := range expected {
		if expected[i] != lastRow[i] {
			t.Errorf("Expected: %s, got: %s", expected[i], lastRow[i])
			return false
		}
	}

	offset, ok := machine.Pool.Search(expected[0])
	var c int
	if !ok {
		fmt.Println("Table not found")
	} else {
		row, err := readRow(int64(offset), TableFile)
		if err != nil {
			fmt.Println("Error finding table: ", err)
		}

		c = DecodeTableCount(row)
		if c != count {
			t.Errorf("Expected: %d, got: %d", count, c)
			return false
		}
	}
	return true
}

func readFile(filename string) []string {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
	}

	fInfo, err := file.Stat()
	buf := make([]byte, fInfo.Size())
	_, err = bufio.NewReader(file).Read(buf)
	if err != nil && err != io.EOF {
		fmt.Println(err)
	}

	// needed to change this janky decoded byte indexing
	decoded := DecodeBytes(buf[len(buf)-21:])
	return decoded
}

func getLastRow(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	var lastRow []string
	var offset int64 = 0

	for {
		_, err = file.Seek(offset, 0)
		if err != nil {
			return nil, fmt.Errorf("error seeking in file: %v", err)
		}

		lengthBytes := make([]byte, 8)
		_, err := file.Read(lengthBytes)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("error reading length: %v", err)
		}

		rowLength := binary.LittleEndian.Uint32(lengthBytes)

		rowData := make([]byte, rowLength)
		_, err = file.Read(rowData)
		if err != nil {
			return nil, fmt.Errorf("error reading row: %v", err)
		}

		decoded := DecodeBytes(rowData)
		lastRow = decoded

		offset += int64(8 + rowLength)
	}

	if lastRow == nil {
		return nil, fmt.Errorf("no rows found in file")
	}

	return lastRow, nil
}

func createDummyTables(t *testing.T) {
	tests := []struct {
		input string
	}{
		{"CREATE TABLE dogs (name varchar, breed varchar);"},
		{"CREATE TABLE coffee (region varchar, brand varchar, roast varchar, size varchar);"},
		{"CREATE TABLE wishlist (name varchar, brand varchar, price varchar);"},
		// {"CREATE INDEX ON wishlist (name, price);"},
		{"CREATE TABLE dogs (name varchar, breed varchar);"},
		{"INSERT INTO coffee VALUES (\"kenya\", \"prodigal\", \"light\", \"65\");"},
		{"INSERT INTO dogs VALUES (\"winnie\", \"cane corso\");"},
		{"INSERT INTO coffee VALUES (\"ethiopia\", \"onyx\", \"light\", \"65\");"},
		{"INSERT INTO coffee VALUES (\"colombia\", \"prodigal\", \"medium\", \"65\");"},
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
		}

		machine := New(comp.Bytecode())
		err = machine.Run()
		if err != nil {
			t.Error("Error running")
		}
	}
}

func createDummyTablesForInsert(t *testing.T) {
	tests := []struct {
		input string
	}{
		{"CREATE TABLE dogs (name varchar, breed varchar);"},
		{"CREATE TABLE wishlist (name varchar, brand varchar, price varchar);"},
		// {"CREATE TABLE coffee (region varchar, brand varchar, roast varchar, size varchar);"},
		// {"INSERT INTO wishlist (\"4060\", \"nvidia\");"},
		// {"INSERT INTO wishlist (\"4080\", \"nvidia\");"},
		// {"INSERT INTO wishlist (\"4070\", \"nvidia\");"},
		// {"INSERT INTO wishlist (\"4070 ti super\", \"nvidia\");"},
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
		}

		machine := New(comp.Bytecode())
		err = machine.Run()
		if err != nil {
			t.Error("Error running")
		}
	}
}

func TestSelect(t *testing.T) {
	createDummyTables(t)
	tests := []struct {
		input string
	}{
		{"SELECT * FROM dogs WHERE breed = \"cane corso\";"},
	}

	for _, tt := range tests {
		program := createParseProgram(tt.input, t)
		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain 1 statements. got=%d", len(program.Statements))
		}

		stmt := program.Statements[0]
		comp := c.New()

		if !testSelect(t, stmt, comp) {
			return
		}

	}

	os.Remove(IdxFile)
	os.Remove(TableFile)
	os.Remove(RowsFile)
}

func testSelect(t *testing.T, stmt ast.Statement, comp *c.Compiler) bool {
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

	for machine.sp > 0 {
		val := machine.pop()
		if v, ok := val.(*code.FoundRow); ok {
			if v.Val[0] != "winnie" {
				t.Errorf("expected: winnie, got: %s", v.Val[0])
				return false
			}
			if v.Val[1] != "cane corso" {
				t.Errorf("expected: cane corso, got: %s", v.Val[1])
				return false
			}
		}
	}
	return true
}

func TestAddIndex(t *testing.T) {
	createDummyTables(t)
	tests := []struct {
		input    string
		tName    string
		idx      []int
		expected []string
		newIdxs  []string
	}{
		// {"CREATE INDEX ON wishlist (name, price);", "wishlist", []int{3, 13}, []string{"true", "true"}},
		{"CREATE INDEX ON coffee (region);", "coffee", []int{3}, []string{"true"}, []string{"colombia", "ethiopia", "kenya"}},
	}

	for _, tt := range tests {
		program := createParseProgram(tt.input, t)
		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain 1 statements. got=%d", len(program.Statements))
		}

		stmt := program.Statements[0]
		comp := c.New()

		if !testAddIndex(t, stmt, comp, tt.tName, tt.idx, tt.expected, tt.newIdxs) {
			return
		}

	}

	os.Remove(IdxFile)
	os.Remove(RowsFile)
	os.Remove(TableFile)
}

func testAddIndex(t *testing.T, stmt ast.Statement, comp *c.Compiler, tName string, idxs []int, expected []string, check []string) bool {
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

	offset, ok := machine.Pool.Search(tName)
	decoded := []string{}
	if !ok {
		fmt.Println("Table not found")
	} else {
		row, err := readRow(int64(offset), TableFile)
		if err != nil {
			fmt.Println("Error finding table: ", err)
		}

		decoded = DecodeBytes(row)
	}

	for i := range idxs {
		idx := idxs[i]
		if decoded[idx] != expected[i] {
			t.Errorf("expected: %s, got: %s", expected, decoded[idx])
			return false
		}
	}

	for i := range check {
		newOffset, ok := machine.Pool.Search(check[i])
		dRow := []string{}
		if !ok {
			fmt.Println("Table not found")
		} else {
			row, err := readRow(int64(newOffset), RowsFile)
			if err != nil {
				fmt.Println("Error finding table: ", err)
			}

			dRow = DecodeBytes(row)
		}

		if dRow[1] != check[i] {
			t.Errorf("expected: %s, got: %s", check[i], dRow[1])
			return false
		}
	}

	return true
}
