package compile

import (
	"encoding/binary"

	"github.com/aidanjjenkins/compiler/ast"
	"github.com/aidanjjenkins/compiler/code"
)

// table row layout
//  len       cell cols...
// | 8 bytes | ???

// table cell layout
// name | type | index | unique | primary key |

const (
	Lengths         = 4
	TotalLengths    = 8
	RowMetaDataSize = 295

	TableMetaDataSize    = 335
	TableMetaDataLengths = 80
	TableIndexSize       = 255
	TableNameSize        = 255
)

type Compiler struct {
	Instructions code.Instructions
	Writes       [][]byte
	constants    []code.Obj
}

func New() *Compiler {
	c := &Compiler{
		Instructions: code.Instructions{},
		Writes:       [][]byte{},
	}

	return c
}

func (c *Compiler) Compile(node ast.Node) error {
	switch node := node.(type) {
	case *ast.Program:
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}
	case *ast.CreateTableStatement:
		tName := &code.TableName{Value: node.TName.Val}
		c.emit(code.OpEncodeStringVal, c.addConstant(tName))
		// need definitions of builtins, find emit should be call,
		for i := range node.Cols {
			cell := &code.ColCell{Name: node.Cols[i], ColType: node.ColTypes[i], Unique: false, Index: false, Pk: false}
			// instruction give is to encode and push to stack
			c.emit(code.OpEncodeTableCell, c.addConstant(cell))
		}
		c.emit(code.OpCreateTable, len(node.Cols)+1)
	case *ast.CreateIndexStatement:
		// for every index given, add constant and emit code to dq, encode, push to stack
		// when done, final instruction will add name to constants, and tell stack to dq all values off the stack to actually execute the add index
		for i := range node.Cols {
			col := &code.Col{Value: node.Cols[i].Val}
			c.emit(code.OpAddIndex, c.addConstant(col))
		}
		table := &code.TableName{Value: node.TName.Val}
		c.emit(code.OpCreateTableIndex, c.addConstant(table))
	case *ast.SelectStatement:
		tName := &code.TableName{Value: node.TName.Val}
		c.emit(code.OpConstant, c.addConstant(tName))
		for i := range node.Condition {
			col := &code.Where{Column: node.Condition[i].CName.Val, Value: node.Condition[i].CIdent}
			c.emit(code.OpWhereCondition, c.addConstant(col))
		}
		c.emit(code.OpSelect, len(node.Condition)+1)
	case *ast.InsertStatement:
		// this shouldnt bne encode isntruction, should be constant
		tName := &code.TableName{Value: node.TName.Val}
		c.emit(code.OpEncodeStringVal, c.addConstant(tName))
		for i := range node.Values {
			col := &code.Col{Value: node.Values[i]}
			c.emit(code.OpEncodeStringVal, c.addConstant(col))
		}
		c.emit(code.OpInsertRow, len(node.Values)+1)
	case *ast.UpdateStatement:
	}
	return nil
}

func encodeString(s string) []byte {
	encodedBytes := []byte(s)
	encodedBytes = append(encodedBytes, 0x00)
	return encodedBytes
}

func encodeBool(b bool) byte {
	if b {
		return 0xFF
	}
	return 0xFD
}

func DecodeBytes(data []byte) []string {
	var result []string

	for i := 0; i < len(data); i++ {
		if data[i] == 0x00 { // Null terminating byte
			break
		} else if data[i] == 0xFF { // True byte
			result = append(result, "true")
		} else if data[i] == 0xFD { // False byte
			result = append(result, "false")
		} else { // String byte
			str := ""
			for j := i; j < len(data) && data[j] != 0x00; j++ {
				str += string(data[j])
				i++
			}
			result = append(result, str)
		}
	}

	return result
}

func createTableRow(node ast.CreateTableStatement) []byte {
	row := []byte{}

	encodedName := encodeString(node.TName.Val)
	row = append(row, encodedName...)

	for i := range node.Cols {
		col := node.Cols[i]
		colType := node.ColTypes[i]

		encodedVal := encodeString(col)
		row = append(row, encodedVal...)

		encodedType := encodeString(colType)
		row = append(row, encodedType...)

		index := encodeBool(false)
		row = append(row, index)

		unique := encodeBool(false)
		row = append(row, unique)

		pk := encodeBool(false)
		row = append(row, pk)
	}

	totalLength := make([]byte, TotalLengths)
	binary.LittleEndian.PutUint32(totalLength, uint32(len(row)))
	row = append(totalLength, row...)

	return row
}

type Bytecode struct {
	Instructions code.Instructions
	constants    []code.Obj
	Writes       [][]byte
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{Instructions: c.Instructions, constants: c.constants, Writes: c.Writes}
}

func (c *Compiler) addConstant(obj code.Obj) int {
	c.constants = append(c.constants, obj)
	return len(c.constants) - 1
}

func (c *Compiler) addInstruction(ins []byte) int {
	posNewInstruction := len(c.Instructions)
	c.Instructions = append(c.Instructions, ins...)

	return posNewInstruction
}

func (c *Compiler) emit(op code.Opcode, operands ...int) int {
	ins := code.Make(op, operands...)
	pos := c.addInstruction(ins)

	return pos
}
