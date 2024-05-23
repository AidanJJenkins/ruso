package code

import "fmt"

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
	OpCreateTable          Opcode = iota
	OpCreateIndex
	OpSelect
	OpInsert
	OpDelete
	OpUpdate
	OpAddTableToTree
)

func Make(op Opcode) []byte {
	ins := []byte{byte(op)}

	return ins
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
	Table []byte
	Cols  [][]byte
	Vals  [][]byte
}

func (s *Select) Type() Object { return SELECT_OBJ }
func (s *Select) Inspect() string {
	return fmt.Sprintf("Table: %s, Cols: %s, Vals: %s", s.Table, s.Cols, s.Vals)
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
