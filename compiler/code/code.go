package code

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type Instructions []byte

type Object string
type Opcode byte

type Obj interface {
	Type() Object
	Inspect() string
}

const (
	SELECT_OBJ                    = "SELECT"
	INSERT_OBJ                    = "INSERT"
	UPDATE_OBJ                    = "UPDATE"
	DELETE_OBJ                    = "DELETE"
	CREATE_TABLE_OBJ              = "CREATE TABLE"
	CREATE_TABLE_INDEX_OBJ        = "CREATE TABLE INDEX"
	ROW_OFFSET_OBJ                = "ROW OFFSET"
	TABLE_NAME                    = "TABLE NAME"
	ADD_INDEX_OBJ                 = "ADD INDEX"
	COL_OBJ                       = "COLUMN"
	WHERE_OBJ                     = "WHERE"
	OpCreateTable          Opcode = iota
	OpCreateIndex
	OpSelect
	OpInsert
	OpDelete
	OpUpdate
	OpAddTableToTree
	OpCall
	OpPop
	OpConstant
	OpCreateTableIndex
	OpEncodeTableCell
	OpEncodeStringVal
	OpAddIndex
	OpInsertRow
	OpTableNameSearch
	OpWhereCondition
)

type Definition struct {
	Name          string
	OperandWidths []int
}

var definitions = map[Opcode]*Definition{
	OpConstant:         {"OpConstant", []int{2}},
	OpCreateTable:      {"OpCreateTable", []int{1}},
	OpEncodeTableCell:  {"OpEncodeTableCell", []int{2}},
	OpAddIndex:         {"OpAddIndex", []int{2}},
	OpPop:              {"OpPop", []int{}},
	OpCreateTableIndex: {"OpCreateTableIndex", []int{2}},
	OpEncodeStringVal:  {"OpEncodeStringVal", []int{2}},
	OpInsertRow:        {"OpInsertRow", []int{1}},
	OpSelect:           {"OpSelect", []int{1}},
	OpWhereCondition:   {"OpWhereCondition", []int{2}},
}

func Make(op Opcode, operands ...int) []byte {
	def, ok := definitions[op]
	if !ok {
		return []byte{}
	}

	instructionLen := 1
	for _, w := range def.OperandWidths {
		instructionLen += w
	}

	instruction := make([]byte, instructionLen)
	instruction[0] = byte(op)

	offset := 1
	for i, o := range operands {
		width := def.OperandWidths[i]
		switch width {
		case 2:
			binary.BigEndian.PutUint16(instruction[offset:], uint16(o))
		case 1:
			instruction[offset] = byte(o)
		}
		offset += width
	}

	return instruction
}

func ReadOperands(def *Definition, ins Instructions) ([]int, int) {
	operands := make([]int, len(def.OperandWidths))
	offset := 0

	for i, width := range def.OperandWidths {
		switch width {
		case 2:
			operands[i] = int(ReadUint16(ins[offset:]))
		case 1:
			operands[i] = int(ReadUint8(ins[offset:]))
		}

		offset += width
	}

	return operands, offset
}

func ReadUint16(ins Instructions) uint16 {
	///
	return binary.BigEndian.Uint16(ins)
}

func ReadUint8(ins Instructions) uint8 {
	return uint8(ins[0])
}

func (ins Instructions) String() string {
	var out bytes.Buffer

	i := 0
	for i < len(ins) {
		def, err := Lookup(ins[i])
		if err != nil {
			fmt.Fprintf(&out, "ERROR: %s\n", err)
			continue
		}

		operands, read := ReadOperands(def, ins[i+1:])

		fmt.Fprintf(&out, "%04d %s\n", i, ins.fmtInstruction(def, operands))

		i += 1 + read
	}

	return out.String()
}

func (ins Instructions) fmtInstruction(def *Definition, operands []int) string {
	operandCount := len(def.OperandWidths)

	if len(operands) != operandCount {
		return fmt.Sprintf("ERROR: operand len %d does not match defined %d\n",
			len(operands), operandCount)
	}

	switch operandCount {
	case 0:
		return def.Name
	case 1:
		return fmt.Sprintf("%s %d", def.Name, operands[0])
	}

	return fmt.Sprintf("ERROR: unhandled operandCount for %s\n", def.Name)
}

func Lookup(op byte) (*Definition, error) {
	def, ok := definitions[Opcode(op)]
	if !ok {
		return nil, fmt.Errorf("opcode %d undefined", op)
	}

	return def, nil
}

type AddIndex struct {
	TName string
	CName string
}

func (a *AddIndex) Type() Object { return ADD_INDEX_OBJ }
func (a *AddIndex) Inspect() string {
	return fmt.Sprintf("Vals: %s, %s", a.TName, a.CName)
}

type TableName struct {
	Value string
}

func (t *TableName) Type() Object { return TABLE_NAME }
func (t *TableName) Inspect() string {
	return fmt.Sprintf("Vals: %s", t.Value)
}

type RowOffset struct {
	Value uint32
}

func (r *RowOffset) Type() Object { return ROW_OFFSET_OBJ }
func (r *RowOffset) Inspect() string {
	return fmt.Sprintf("Vals: %d", r.Value)
}

type Select struct {
	Table string
}

func (s *Select) Type() Object { return SELECT_OBJ }
func (s *Select) Inspect() string {
	return fmt.Sprintf("Table name: %s", s.Table)
}

type Insert struct {
	Table   []byte
	Vals    []byte
	ValLens []int
	RowLen  int
}

func (i *Insert) Type() Object { return INSERT_OBJ }
func (i *Insert) Inspect() string {
	return fmt.Sprintf("Table: %s,Vals: %s", i.Table, i.Vals)
}

type CreateTable struct {
	Table []byte
	Cols  []byte
	Types []byte
}

func (ct *CreateTable) Type() Object { return CREATE_TABLE_OBJ }
func (ct *CreateTable) Inspect() string {
	return fmt.Sprintf("Table: %s,Cols: %s, Vals: %s", ct.Table, ct.Cols, ct.Types)
}

type CreateTableIndex struct {
	Table []byte
	Cols  [][]byte
}

func (cti *CreateTableIndex) Type() Object { return CREATE_TABLE_OBJ }
func (cti *CreateTableIndex) Inspect() string {
	return fmt.Sprintf("Table: %s,Cols: %s", cti.Table, cti.Cols)
}

type Update struct {
	Table []byte
	Cols  [][]byte
	Vals  [][]byte
	Where [][]byte
}

func (u *Update) Type() Object { return UPDATE_OBJ }
func (u *Update) Inspect() string {
	return fmt.Sprintf("Table: %s,Cols: %s, Vals: %s, Where: %s", u.Table, u.Cols, u.Vals, u.Where)
}

type Delete struct {
	Table []byte
	Cols  [][]byte
	Vals  [][]byte
}

func (d *Delete) Type() Object { return DELETE_OBJ }
func (d *Delete) Inspect() string {
	return fmt.Sprintf("Table: %s,Cols: %s, Vals: %s", d.Table, d.Cols, d.Vals)
}

type Col struct {
	Value string
}

func (c *Col) Type() Object { return COL_OBJ }
func (c *Col) Inspect() string {
	return fmt.Sprintf("Table: %s,Cols: ", c.Value)
}

type ColCell struct {
	Name    string
	ColType string
	Index   bool
	Unique  bool
	Pk      bool
}

func (c *ColCell) Type() Object { return COL_OBJ }
func (c *ColCell) Inspect() string {
	return fmt.Sprintf("Col name: %s,: , Col type: %s, Index: %t, Unique: %t Primay Key: %t", c.Name, c.ColType, c.Index, c.Unique, c.Pk)
}

type Where struct {
	Column string
	Value  string
}

func (w *Where) Type() Object { return WHERE_OBJ }
func (w *Where) Inspect() string {
	return fmt.Sprintf("Column: %s, Value: %s ", w.Column, w.Value)
}
