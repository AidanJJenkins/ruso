package compile

import (
	"encoding/binary"
	"fmt"

	"github.com/aidanjjenkins/compiler/ast"
	"github.com/aidanjjenkins/compiler/code"
)

const (
	Lengths         = 4
	RowMetaDataSize = 295

	TableMetaDataSize    = 335
	TableMetaDataLengths = 80
	TableIndexSize       = 255
	TableNameSize        = 255
)

type Compiler struct {
	Instructions code.Instructions
}

func New() *Compiler {
	c := &Compiler{
		Instructions: code.Instructions{},
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
		ins := code.Make(code.OpCreateTable)
		m := createTableMetaRow(*node)
		ins = append(ins, m...)

		c.Instructions = append(c.Instructions, ins...)
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
		ins := code.Make(code.OpInsert)
		ins = append(ins, createUpdateInstructions(*node)...)

		c.Instructions = append(c.Instructions, ins...)
	}
	return nil
}

func createTableMetaRow(node ast.CreateTableStatement) []byte {
	row := []byte{}
	metaData := make([]byte, TableMetaDataSize)
	lengths := make([]byte, TableMetaDataLengths)

	n := make([]byte, TableNameSize)
	copy(n, []byte(node.TName.Val))
	row = append(row, n...)

	offset := 0
	for i := range node.Cols {
		col := []byte(node.Cols[i])
		typ := []byte(node.ColTypes[i])

		row = append(row, col...)
		row = append(row, typ...)

		cLLen := make([]byte, Lengths)
		tLLen := make([]byte, Lengths)
		binary.LittleEndian.PutUint32(cLLen, uint32(len(col)))
		binary.LittleEndian.PutUint32(tLLen, uint32(len(typ)))

		lengths = append(lengths, cLLen...)
		lengths = append(lengths, tLLen...)

		copy(lengths[offset:offset+4], cLLen)
		copy(lengths[offset+4:offset+8], tLLen)

		offset += 8
	}

	copy(metaData[:255], make([]byte, 255))
	copy(metaData[255:], lengths)

	metaData = append(metaData, row...)

	return metaData
}

func AccessFirstByte(ins []byte) int {
	return int(ins[0])
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
		v := []byte(node.Values[i])
		vLen := make([]byte, Lengths)
		binary.LittleEndian.PutUint32(vLen, uint32(len(v)))

		c := []byte(node.Cols[i].Val)
		cLen := make([]byte, Lengths)
		binary.LittleEndian.PutUint32(cLen, uint32(len(c)))

		cAndV = append(cAndV, vLen...)
		cAndV = append(cAndV, v...)
		cAndV = append(cAndV, cLen...)
		cAndV = append(cAndV, c...)
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

type Bytecode struct {
	Instructions code.Instructions
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{Instructions: c.Instructions}
}
