package compile

import (
	"github.com/aidanjjenkins/compiler/ast"
	"github.com/aidanjjenkins/compiler/code"
	o "github.com/aidanjjenkins/compiler/object"
)

type Compiler struct {
	Instructions code.Instructions
	Constants    []o.Obj
}

func New() *Compiler {
	c := &Compiler{
		Instructions: code.Instructions{},
		Constants:    []o.Obj{},
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
		tName := &o.TableName{Value: node.TName.Val}
		c.emit(code.OpEncodeStringVal, c.addConstant(tName))
		for i := range node.Cols {
			cell := &o.ColCell{Name: node.Cols[i], ColType: node.ColTypes[i], Unique: false, Index: false, Pk: false}
			c.emit(code.OpEncodeTableCell, c.addConstant(cell))
		}
		c.emit(code.OpCreateTable, len(node.Cols)+1)
	case *ast.CreateIndexStatement:
		for i := range node.Cols {
			col := &o.Col{Value: node.Cols[i].Val}
			c.emit(code.OpAddIndex, c.addConstant(col))
		}
		table := &o.TableName{Value: node.TName.Val}
		c.emit(code.OpCreateTableIndex, c.addConstant(table))
	case *ast.SelectStatement:
		tName := &o.TableName{Value: node.TName.Val}
		c.emit(code.OpTableNameSearch, c.addConstant(tName))
		for i := range node.Condition {
			col := &o.Where{Column: node.Condition[i].CName.Val, Value: node.Condition[i].CIdent}
			c.emit(code.OpWhereCondition, c.addConstant(col))
		}
		c.emit(code.OpSelect, len(node.Condition)+1)
	case *ast.InsertStatement:
		tName := &o.TableName{Value: node.TName.Val}
		c.emit(code.OpEncodeStringVal, c.addConstant(tName))

		numVals := 0
		if node.Cols != nil {
			c.Compile(node.Cols)
			numVals += len(node.Cols.Values)
		}
		c.Compile(node.Vals)
		numVals += len(node.Vals.Values)
		c.emit(code.OpInsertRow, numVals+1)
	case *ast.InsertVals:
		// each type literal should have its own encod<type> instruction, that why type checking can be done in the
		// vertual machine
		for i := range node.Values {
			switch v := node.Values[i].(type) {
			case *ast.StringLiteral:
				col := &o.Col{Value: v.Value}
				c.emit(code.OpEncodeStringVal, c.addConstant(col))
			}
		}
	case *ast.InsertCols:
		for i := range node.Values {
			col := &o.Col{Value: node.Values[i].Val}
			c.emit(code.OpCheckCol, c.addConstant(col))
		}
		//emit create tableInfo obj
	case *ast.UpdateStatement:
	}
	return nil
}

type Bytecode struct {
	Instructions code.Instructions
	Constants    []o.Obj
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{Instructions: c.Instructions, Constants: c.Constants}
}

func (c *Compiler) addConstant(obj o.Obj) int {
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
