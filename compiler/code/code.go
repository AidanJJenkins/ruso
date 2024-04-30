package code

import "fmt"

type ObjectType string

type Obj interface {
	Type() ObjectType
	Inspect() string
}

const (
	SELECT_OBJ             = "SELECT"
	INSERT_OBJ             = "INSERT"
	UPDATE_OBJ             = "UPDATE"
	DELETE_OBJ             = "DELETE"
	CREATE_TABLE_OBJ       = "CREATE TABLE"
	CREATE_TABLE_INDEX_OBJ = "CREATE TABLE INDEX"
)

type Select struct {
	Table []byte
	Cols  [][]byte
	Vals  [][]byte
}

func (s *Select) Type() ObjectType { return SELECT_OBJ }
func (s *Select) Inspect() string {
	return fmt.Sprintf("Table: %s, Cols: %s, Vals: %s", s.Table, s.Cols, s.Vals)
}

type Insert struct {
	Table []byte
	Vals  [][]byte
}

func (i *Insert) Type() ObjectType { return INSERT_OBJ }
func (i *Insert) Inspect() string {
	return fmt.Sprintf("Table: %s,Vals: %s", i.Table, i.Vals)
}

type Update struct {
	Table []byte
	Cols  [][]byte
	Vals  [][]byte
	Where [][]byte
}

func (u *Update) Type() ObjectType { return UPDATE_OBJ }
func (u *Update) Inspect() string {
	return fmt.Sprintf("Table: %s,Cols: %s, Vals: %s, Where: %s", u.Table, u.Cols, u.Vals, u.Where)
}

type Delete struct {
	Table []byte
	Cols  [][]byte
	Vals  [][]byte
}

func (d *Delete) Type() ObjectType { return DELETE_OBJ }
func (d *Delete) Inspect() string {
	return fmt.Sprintf("Table: %s,Cols: %s, Vals: %s", d.Table, d.Cols, d.Vals)
}

type CreateTable struct {
	Table []byte
	Cols  [][]byte
	Types [][]byte
}

func (ct *CreateTable) Type() ObjectType { return CREATE_TABLE_OBJ }
func (ct *CreateTable) Inspect() string {
	return fmt.Sprintf("Table: %s,Cols: %s, Vals: %s", ct.Table, ct.Cols, ct.Types)
}

type CreateTableIndex struct {
	Table []byte
	Cols  [][]byte
}

func (cti *CreateTableIndex) Type() ObjectType { return CREATE_TABLE_OBJ }
func (cti *CreateTableIndex) Inspect() string {
	return fmt.Sprintf("Table: %s,Cols: %s", cti.Table, cti.Cols)
}
