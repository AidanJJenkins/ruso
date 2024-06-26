package compile

import (
	"github.com/aidanjjenkins/compiler/ast"
	"github.com/aidanjjenkins/compiler/code"
)

type Compiler struct {
	Instructions code.Instructions
	Constants    []code.Obj
}

func New() *Compiler {
	c := &Compiler{
		Instructions: code.Instructions{},
		Constants:    []code.Obj{},
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
		for i := range node.Cols {
			cell := &code.ColCell{Name: node.Cols[i], ColType: node.ColTypes[i], Unique: false, Index: false, Pk: false}
			c.emit(code.OpEncodeTableCell, c.addConstant(cell))
		}
		c.emit(code.OpCreateTable, len(node.Cols)+1)
	case *ast.CreateIndexStatement:
		for i := range node.Cols {
			col := &code.Col{Value: node.Cols[i].Val}
			c.emit(code.OpConstant, c.addConstant(col))
		}
		table := &code.TableName{Value: node.TName.Val}
		c.emit(code.OpCreateTableIndex, c.addConstant(table))
	case *ast.SelectStatement:
		tName := &code.TableName{Value: node.TName.Val}
		c.emit(code.OpTableNameSearch, c.addConstant(tName))
		for i := range node.Condition {
			col := &code.Where{Column: node.Condition[i].CName.Val, Value: node.Condition[i].CIdent}
			c.emit(code.OpWhereCondition, c.addConstant(col))
		}
		c.emit(code.OpSelect, len(node.Condition)+1)
		// case *ast.InsertStatement:
		// 	tName := &code.TableName{Value: node.TName.Val}
		// 	c.emit(code.OpEncodeStringVal, c.addConstant(tName))
		// 	for i := range node.Values {
		// 		col := &code.Col{Value: node.Values[i]}
		// 		c.emit(code.OpEncodeStringVal, c.addConstant(col))
		// 	}
		// 	c.emit(code.OpInsertRow, len(node.Values)+1)
	case *ast.InsertStatement:
		tName := &code.TableName{Value: node.TName.Val}
		c.emit(code.OpTableInfo, c.addConstant(tName))

		numVals := 0
		if node.Cols != nil {
			c.Compile(node.Cols)
			numVals += len(node.Cols.Values)
		}
		c.Compile(node.Vals)
		numVals += len(node.Vals.Values)
		c.emit(code.OpInsert)
	case *ast.InsertVals:
		for i := range node.Values {
			switch v := node.Values[i].(type) {
			case *ast.StringLiteral:
				col := &code.Col{Value: v.Value}
				c.emit(code.OpValInfo, c.addConstant(col))
			}
		}
	case *ast.InsertCols:
		for i := range node.Values {
			col := &code.Col{Value: node.Values[i].Val}
			c.emit(code.OpColInfo, c.addConstant(col))
		}
	case *ast.UpdateStatement:
	}
	return nil
}

type Bytecode struct {
	Instructions code.Instructions
	Constants    []code.Obj
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{Instructions: c.Instructions, Constants: c.Constants}
}

func (c *Compiler) addConstant(obj code.Obj) int {
	c.Constants = append(c.Constants, obj)
	return len(c.Constants) - 1
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
