package object

import "fmt"

type Object string

type Obj interface {
	Type() Object
	Inspect() string
}

const (
	SELECT_OBJ             = "SELECT"
	INSERT_OBJ             = "INSERT"
	UPDATE_OBJ             = "UPDATE"
	DELETE_OBJ             = "DELETE"
	CREATE_TABLE_OBJ       = "CREATE TABLE"
	CREATE_TABLE_INDEX_OBJ = "CREATE TABLE INDEX"
	ROW_OFFSET_OBJ         = "ROW OFFSET"
	TABLE_NAME             = "TABLE NAME"
	ADD_INDEX_OBJ          = "ADD INDEX"
	COL_OBJ                = "COLUMN"
	WHERE_OBJ              = "WHERE"
	ENCODED_VAL            = "ENCODED_VAL"
	FOUND_ROW              = "FOUND_ROW"
	CONSTANT_VAL           = "CONSTANT_VAL"
)

//	type AddIndex struct {
//		TName string
//		CName string
//	}
//
// func (a *AddIndex) Type() Object { return ADD_INDEX_OBJ }
//
//	func (a *AddIndex) Inspect() string {
//		return fmt.Sprintf("Vals: %s, %s", a.TName, a.CName)
//	}
type TableName struct {
	Value string
}

func (t *TableName) Type() Object { return TABLE_NAME }
func (t *TableName) Inspect() string {
	return fmt.Sprintf("Vals: %s", t.Value)
}

//	type RowOffset struct {
//		Value uint32
//	}
//
// func (r *RowOffset) Type() Object { return ROW_OFFSET_OBJ }
//
//	func (r *RowOffset) Inspect() string {
//		return fmt.Sprintf("Vals: %d", r.Value)
//	}
//
//	type Select struct {
//		Table string
//	}
//
// func (s *Select) Type() Object { return SELECT_OBJ }
//
//	func (s *Select) Inspect() string {
//		return fmt.Sprintf("Table name: %s", s.Table)
//	}
//
//	type Insert struct {
//		Table   []byte
//		Vals    []byte
//		ValLens []int
//		RowLen  int
//	}
//
// func (i *Insert) Type() Object { return INSERT_OBJ }
//
//	func (i *Insert) Inspect() string {
//		return fmt.Sprintf("Table: %s,Vals: %s", i.Table, i.Vals)
//	}
//
//	type CreateTable struct {
//		Table []byte
//		Cols  []byte
//		Types []byte
//	}
//
// func (ct *CreateTable) Type() Object { return CREATE_TABLE_OBJ }
//
//	func (ct *CreateTable) Inspect() string {
//		return fmt.Sprintf("Table: %s,Cols: %s, Vals: %s", ct.Table, ct.Cols, ct.Types)
//	}
//
//	type CreateTableIndex struct {
//		Table []byte
//		Cols  [][]byte
//	}
//
// func (cti *CreateTableIndex) Type() Object { return CREATE_TABLE_OBJ }
//
//	func (cti *CreateTableIndex) Inspect() string {
//		return fmt.Sprintf("Table: %s,Cols: %s", cti.Table, cti.Cols)
//	}
//
//	type Update struct {
//		Table []byte
//		Cols  [][]byte
//		Vals  [][]byte
//		Where [][]byte
//	}
//
// func (u *Update) Type() Object { return UPDATE_OBJ }
//
//	func (u *Update) Inspect() string {
//		return fmt.Sprintf("Table: %s,Cols: %s, Vals: %s, Where: %s", u.Table, u.Cols, u.Vals, u.Where)
//	}
//
//	type Delete struct {
//		Table []byte
//		Cols  [][]byte
//		Vals  [][]byte
//	}
//
// func (d *Delete) Type() Object { return DELETE_OBJ }
//
//	func (d *Delete) Inspect() string {
//		return fmt.Sprintf("Table: %s,Cols: %s, Vals: %s", d.Table, d.Cols, d.Vals)
//	}
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

type EncodedVal struct {
	Val []byte
}

func (ev *EncodedVal) Type() Object { return ENCODED_VAL }
func (ev *EncodedVal) Inspect() string {
	return fmt.Sprintf("Encoded bytes: %s,:", ev.Val)
}

type Constant struct {
	Val string
}

func (c *Constant) Type() Object { return ENCODED_VAL }
func (c *Constant) Inspect() string {
	return fmt.Sprintf("Encoded bytes: %s,:", c.Val)
}

type FoundRow struct {
	Val []string
}

func (fr *FoundRow) Type() Object { return FOUND_ROW }
func (fr *FoundRow) Inspect() string {
	return fmt.Sprintf("Found row: %s,:", fr.Val)
}
