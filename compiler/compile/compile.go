package compile

import (
	"encoding/binary"
	"fmt"

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
		makeTableInstruction := code.Make(code.OpCreateTable)
		addTableToTree := code.Make(code.OpAddTableToTree)

		m := createTableRow(*node)
		c.Writes = append(c.Writes, m)

		c.Instructions = append(c.Instructions, makeTableInstruction...)
		c.Instructions = append(c.Instructions, addTableToTree...)
	case *ast.CreateIndexStatement:
		ins := code.Make(code.OpCreateIndex)
		ins = append(ins, createIndexInstructions(*node)...)

		c.Instructions = append(c.Instructions, ins...)
	case *ast.SelectStatement:
		ins := code.Make(code.OpSelect)
		ins = append(ins, createSelectInstructions(*node)...)

		c.Instructions = append(c.Instructions, ins...)
	case *ast.InsertStatement:
		ins := code.Make(code.OpInsert)
		ins = append(ins, createInsertInstructions(*node)...)

		c.Instructions = append(c.Instructions, ins...)
	case *ast.UpdateStatement:
		ins := code.Make(code.OpUpdate)
		ins = append(ins, createUpdateInstructions(*node)...)

		c.Instructions = append(c.Instructions, ins...)
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

func AccessFirstByte(ins []byte) int {
	return int(ins[0])
}

func AccessTotalRowLength(ins []byte) uint32 {
	b := ins[1:9]
	l := binary.LittleEndian.Uint32(b)

	return l
}

func AccessTableMetaDataIndex(ins []byte) {
	fmt.Println("MetaData Index", ins[1:1+TableIndexSize])
}

func AccessTableMetaDataLengths(ins []byte) []int {
	lens := ins[TableIndexSize+1 : TableIndexSize+TableMetaDataLengths]
	l := []int{}

	for i := range lens {
		if lens[i] != 0 {
			l = append(l, int(lens[i]))
		}
	}

	return l
}

func AccessTableNameBytes(ins []byte) string {
	return string(ins[TableMetaDataSize+1 : TableMetaDataSize+TableNameSize])
}

func AccessTableRowInfoBytes(ins []byte, lengths []int) []string {
	vals := []string{}
	for _, l := range lengths {
		if l < len(ins) {
			value := ins[:l]
			vals = append(vals, string(value))
			ins = ins[l:]
		}
	}
	vals = append(vals, string(ins))

	return vals
}

func createIndexInstructions(node ast.CreateIndexStatement) []byte {
	n := make([]byte, TableNameSize)
	copy(n, []byte(node.TName.Val))

	for i := range node.Cols {
		col := []byte(node.Cols[i].Val)

		cLLen := make([]byte, Lengths)
		binary.LittleEndian.PutUint32(cLLen, uint32(len(col)))

		n = append(n, cLLen...)
		n = append(n, col...)
	}

	return n
}

func AccessIndexName(ins []byte) string {
	return string(ins[1 : TableNameSize+1])
}

func AccessIndexVals(ins []byte) []string {
	vals := []string{}

	data := ins[TableIndexSize+1:]
	for len(data) > 0 {
		colLength := binary.LittleEndian.Uint32(data[:Lengths])
		data = data[Lengths:]

		colValue := string(data[:colLength])
		data = data[colLength:]

		vals = append(vals, colValue)
	}

	return vals
}

func createSelectInstructions(node ast.SelectStatement) []byte {
	ins := []byte{}
	tName := []byte(node.TName.Val)

	tLen := make([]byte, Lengths)
	binary.LittleEndian.PutUint32(tLen, uint32(len(tName)))

	ins = append(ins, tLen...)
	ins = append(ins, tName...)

	for i := range node.Condition {
		col := []byte(node.Condition[i].CName.Val)
		cLen := make([]byte, Lengths)
		binary.LittleEndian.PutUint32(cLen, uint32(len(col)))

		ident := []byte(node.Condition[i].CIdent)
		iLen := make([]byte, Lengths)
		binary.LittleEndian.PutUint32(iLen, uint32(len(ident)))

		ins = append(ins, cLen...)
		ins = append(ins, col...)
		ins = append(ins, iLen...)
		ins = append(ins, ident...)
	}

	return ins
}

func AccessSelectValues(ins []byte) []string {
	vals := []string{}
	data := ins[1:]
	for len(data) > 0 {
		colLength := binary.LittleEndian.Uint32(data[:Lengths])
		data = data[Lengths:]

		colValue := string(data[:colLength])
		data = data[colLength:]

		vals = append(vals, colValue)
	}

	return vals
}

func createInsertInstructions(node ast.InsertStatement) []byte {
	ins := []byte{}
	totalRowlen := make([]byte, TotalLengths)
	ins = append(ins, totalRowlen...)
	tName := []byte(node.TName.Val)

	tLen := make([]byte, Lengths)
	binary.LittleEndian.PutUint32(tLen, uint32(len(tName)))

	ins = append(ins, tLen...)
	ins = append(ins, tName...)

	for i := range node.Values {
		col := []byte(node.Values[i])
		cLen := make([]byte, Lengths)
		binary.LittleEndian.PutUint32(cLen, uint32(len(col)))

		ins = append(ins, cLen...)
		ins = append(ins, col...)
	}
	binary.LittleEndian.PutUint32(ins[:TotalLengths], uint32(len(ins)))

	return ins
}

func createUpdateInstructions(node ast.UpdateStatement) []byte {
	ins := []byte{}
	tName := []byte(node.TName.Val)

	tLen := make([]byte, Lengths)
	binary.LittleEndian.PutUint32(tLen, uint32(len(tName)))

	ins = append(ins, tLen...)
	ins = append(ins, tName...)

	cAndV := []byte{}
	for i := range node.Values {

		c := []byte(node.Cols[i].Val)
		cLen := make([]byte, Lengths)
		binary.LittleEndian.PutUint32(cLen, uint32(len(c)))

		cAndV = append(cAndV, cLen...)
		cAndV = append(cAndV, c...)

		v := []byte(node.Values[i])
		vLen := make([]byte, Lengths)
		binary.LittleEndian.PutUint32(vLen, uint32(len(v)))

		cAndV = append(cAndV, vLen...)
		cAndV = append(cAndV, v...)

	}

	colAndValLen := make([]byte, Lengths)
	binary.LittleEndian.PutUint32(colAndValLen, uint32(len(cAndV)))
	ins = append(ins, colAndValLen...)
	ins = append(ins, cAndV...)

	conditions := []byte{}
	for j := range node.Condition {
		col := []byte(node.Condition[j].CName.Val)
		cLen := make([]byte, Lengths)
		binary.LittleEndian.PutUint32(cLen, uint32(len(col)))

		ident := []byte(node.Condition[j].CIdent)
		iLen := make([]byte, Lengths)
		binary.LittleEndian.PutUint32(iLen, uint32(len(ident)))

		conditions = append(conditions, cLen...)
		conditions = append(conditions, col...)
		conditions = append(conditions, iLen...)
		conditions = append(conditions, ident...)
	}

	ins = append(ins, conditions...)

	return ins
}

func AccessUpdateName(ins []byte) (string, []byte) {
	ins = ins[1:]
	nLen := binary.LittleEndian.Uint32(ins[:Lengths])
	ins = ins[Lengths:]

	name := string(ins[:nLen])
	ins = ins[nLen:]

	return name, ins
}

func splitSlice(bytes []byte) ([]byte, []byte) {
	length := binary.LittleEndian.Uint32(bytes[:Lengths])

	part1 := bytes[Lengths : Lengths+length]
	part2 := bytes[length+Lengths:]

	return part1, part2
}

func AccessUpdateValues(data []byte) []string {
	vals := []string{}
	for len(data) > 0 {
		colLength := binary.LittleEndian.Uint32(data[:Lengths])
		data = data[Lengths:]

		colValue := string(data[:colLength])
		data = data[colLength:]

		vals = append(vals, colValue)
	}

	return vals
}

type Bytecode struct {
	Instructions code.Instructions
	Writes       [][]byte
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{Instructions: c.Instructions, Writes: c.Writes}
}
