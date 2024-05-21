package vm

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"

	tree "github.com/aidanjjenkins/bplustree"
	"github.com/aidanjjenkins/compiler/code"
	c "github.com/aidanjjenkins/compiler/compile"
)

const TableFile string = "tables.db"
const IdxFile string = "index.db"
const RowsFile string = "rows.db"

type Tbls struct {
	Cols map[string]string // name -> type
	Idx  []string
}

type Pool struct {
	db tree.Pager
}

func newPool() *Pool {
	p := &Pool{}
	p.db.Path = IdxFile
	err := p.db.Open()
	if err != nil {
		fmt.Println("err: ", err)
	}
	return p
}

type VM struct {
	Pool         *Pool
	Instructions code.Instructions
	Info         map[string]*Tbls
}

func New(bytecode *c.Bytecode, info map[string]*Tbls) *VM {
	vm := VM{Pool: newPool(), Instructions: bytecode.Instructions, Info: info}

	return &vm
}

func (p *Pool) Add(key []byte, val uint32) {
	offsetBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(offsetBytes, val)
	p.db.Set(key, offsetBytes)
}

func (p *Pool) Search(key string) (uint32, bool) {
	bytes, found := p.db.Get(key)
	if found {
		value := binary.LittleEndian.Uint32(bytes)
		return value, true
	} else {
		fmt.Println(">>> Value not found")
		return 0, false
	}
}

type VMError struct {
	Message string
}

func (e *VMError) Error() string {
	return e.Message
}

func NewError(message string) error {
	return &VMError{Message: message}
}

func (vm *VM) Run() error {
	ins := vm.Instructions

	op := code.Opcode(c.AccessFirstByte(ins))
	switch op {
	case code.OpCreateTable:
		vm.addTable(vm.Instructions)
	case code.OpCreateIndex:
		vm.addIndex(vm.Instructions)
	case code.OpSelect:
		// vm.search(vm.Instructions)
	case code.OpInsert:
		vm.add(vm.Instructions)
	case code.OpUpdate:
		update(vm.Instructions)
	}
	return nil
}

func (vm *VM) addTable(ins []byte) {
	if vm.Info != nil {
		name := AccessTableNameBytes(ins[1:])
		if vm.Info[name] != nil {
			fmt.Println("Table name already exists.")
			return
		}
	}

	offset, err := writeRow(ins[1:], TableFile)

	if err != nil {
		fmt.Println("error: ", err)
	}

	name := AccessTableNameBytes(ins[1:])
	lens := c.AccessTableMetaDataLengths(ins[1:])
	vals := c.AccessTableRowInfoBytes(ins[9+c.TableMetaDataSize+255:], lens)

	infoCols := make(map[string]string)
	i := 0
	for i < len(vals) {
		n := vals[i]
		t := vals[i+1]

		infoCols[n] = t
		i += 2
	}

	tbls := &Tbls{Cols: infoCols, Idx: nil}
	vm.Info[name] = tbls

	vm.Pool.Add([]byte(name), offset)
}

func (vm *VM) FindTable(name string) []byte {
	offset, ok := vm.Pool.Search(name)
	if !ok {
		fmt.Println("key not found")
		return nil
	} else {
		row, err := readRow(int64(offset), TableFile)
		if err != nil {
			fmt.Println("Error finding table: ", err)
		}

		return row
	}
}

func writeRow(data []byte, filename string) (uint32, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	fileLen := fileInfo.Size()

	_, err = file.Write(data)
	if err != nil {
		return 0, err
	}

	return uint32(fileLen), nil
}

func updateSameLength(offset uint32, data []byte) error {
	file, err := os.OpenFile(TableFile, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteAt(data, int64(offset+8))
	if err != nil {
		return err
	}

	return nil
}

func readRow(offset int64, filename string) ([]byte, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Read the first 8 bytes to get the total length
	lengthBytes := make([]byte, 8)
	_, err = file.ReadAt(lengthBytes, int64(offset))
	if err != nil {
		return nil, err
	}

	length := binary.LittleEndian.Uint64(lengthBytes)
	// Calculate the end offset

	bytes := make([]byte, length)
	_, err = file.ReadAt(bytes, offset+8)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func accessTotalRowLength(ins []byte) uint32 {
	b := ins[:9]
	l := binary.LittleEndian.Uint32(b)

	return l
}

func accessTableMetaDataLengths(ins []byte) []int {
	lens := ins[c.TableIndexSize : c.TableIndexSize+c.TableMetaDataLengths]
	l := []int{}

	for i := range lens {
		if lens[i] != 0 {
			l = append(l, int(lens[i]))
		}
	}

	return l
}

func cleanse(substring []byte) string {
	endIndex := bytes.IndexAny(substring, "\x00 \t\n\r")
	if endIndex == -1 {
		endIndex = len(substring)
	}

	trimmedSubstring := substring[:endIndex]

	return string(trimmedSubstring)
}

// for the insertion
func AccessTableNameBytes(ins []byte) string {
	substring := ins[c.TableMetaDataSize+8 : c.TableMetaDataSize+c.TableNameSize]
	name := cleanse(substring)
	return name
}

// for actually find the name when searching for row
func GetTableNameBytes(ins []byte) string {
	substring := ins[c.TableMetaDataSize : c.TableMetaDataSize+c.TableNameSize]
	name := cleanse(substring)
	return name
}

func accessTableRowInfoBytes(row []byte, lengths []int) []string {
	ins := row[c.TableMetaDataSize+c.TableNameSize:]
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

func AccessIndexName(ins []byte) []byte {
	return ins[1 : c.TableNameSize+1]
}

func (vm *VM) addIndex(ins []byte) {
	vals := c.AccessIndexVals(ins)
	idx := vals[0]
	t := AccessIndexName(ins)
	tName := cleanse(t)

	offset, ok := vm.Pool.Search(tName)
	if !ok {
		fmt.Println("key not found")
		return
	} else {
		updateSameLength(offset, []byte(idx))

	}

}

// func (vm *VM) search(ins []byte) []byte {
// vals := c.AccessSelectValues(ins)
// fmt.Println("vals", vals)
//
// tableName := vals[0]
// col := vals[1]
// val := vals[2]

// offset, ok := vm.Pool.Search(val)
// if !ok {
// 	fmt.Println("key not found")
// 	return nil
// } else {
// 	row, err := readRow(int64(offset))
// 	if err != nil {
// 		fmt.Println("error: ", err)
// 	}
//
// return nil
// }

func (vm *VM) add(ins []byte) {
	values := c.AccessSelectValues(ins)
	name := values[0]
	info, ok := vm.Info[name]
	if !ok {
		fmt.Println("no table with that name")
	}

	cols := []string{}
	for k := range info.Cols {
		cols = append(cols, k)
	}
	if len(cols) < len(values[1:]) {
		fmt.Println("Too many values for table")
		return
	}
	idxOfIdx := 0
	if len(info.Idx) > 0 {
		for i := range cols {
			if cols[i] == info.Idx[0] {
				idxOfIdx = i
			}
		}
		offset, err := writeRow(ins[1:], RowsFile)
		if err != nil {
			fmt.Println("Error writing row: ", err)
			return
		}
		vm.Pool.Add([]byte(values[idxOfIdx]), offset)
		return
	}
	err := writeRowWithoutIndex(ins[1:], RowsFile)
	if err != nil {
		fmt.Println("Error writing row: ", err)
		return
	}
}

func writeRowWithoutIndex(data []byte, filename string) error {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		return err
	}

	return nil
}
func update(ins []byte) {
	fmt.Println("need to implement update")
	fmt.Println("instructions: ", ins)
}
