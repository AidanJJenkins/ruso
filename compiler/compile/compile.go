package compile

import (
	"github.com/aidanjjenkins/compiler/ast"
	"github.com/aidanjjenkins/compiler/code"
)

func Compile(node ast.Node) code.Obj {
	switch node := node.(type) {
	case *ast.InsertStatement:
		obj := &code.Insert{Table: []byte(node.TName.Val)}
		for i := range node.Values {
			obj.Vals = append(obj.Vals, []byte(node.Values[i]))
		}

		return obj
	case *ast.SelectStatement:
		obj := &code.Select{Table: []byte(node.TName.Val)}
		for i := range node.Condition {
			obj.Cols = append(obj.Cols, []byte(node.Condition[i].CName.Val))
			obj.Vals = append(obj.Vals, []byte(node.Condition[i].CIdent.Val))
		}

		return obj
	case *ast.CreateTableStatement:
		obj := &code.CreateTable{Table: []byte(node.TName.Val)}
		for i := range node.Cols {
			obj.Cols = append(obj.Cols, []byte(node.Cols[i]))
			obj.Types = append(obj.Types, []byte(node.ColTypes[i]))
		}

		return obj
	}

	return nil
}
