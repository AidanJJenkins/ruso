package compile

import (
	"encoding/binary"

	"github.com/aidanjjenkins/compiler/ast"
	"github.com/aidanjjenkins/compiler/code"
)

const (
	TableNameSize   = 255
	Lengths         = 4
	RowMetaDataSize = 295

	TableIndexSize    = 255
	TableMetaDataSize = 335
)

// create a compiler struct, when compling add the object as an instruction to an objects slice
type Compiler struct {
	Instructions code.Instructions
}

func New() *Compiler {
	c := &Compiler{
		Instructions: code.Instructions{},
	}

	return c
}

// compile creates the full row including metadata, make the first bite the opcode, emits to the instructions field

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
		key := make([]byte, 255)
		copy(key[:len([]byte(node.TName.Val))], []byte(node.TName.Val))
		ins = append(ins, key...)

		createTableMetaRow(*node)
		// fmt.Println("inst", ins)
		// c.Instructions = append(c.Instructions, ins...)
	}
	return nil
}

func createTableMetaRow(node ast.CreateTableStatement) []byte {
	row := []byte{}
	metaData := make([]byte, TableMetaDataSize)
	lengths := make([]byte, 80)

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
