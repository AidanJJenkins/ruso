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
const StackSize = 2040

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
	constants    []code.Obj
	Writes       [][]byte
	Stack        []code.Obj
	sp           int
}

func New(bytecode *c.Bytecode) *VM {
	vm := VM{
		Pool:         newPool(),
		Instructions: bytecode.Instructions,
		Writes:       bytecode.Writes,
		Stack:        make([]code.Obj, StackSize),
		sp:           0,
	}

	return &vm
}

func (p *Pool) Add(key string, val uint32) {
	k := []byte(key)
	offsetBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(offsetBytes, val)
	p.db.Set(k, offsetBytes)
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

func (vm *VM) push(obj code.Obj) error {
	if vm.sp >= StackSize {
		return fmt.Errorf("Stack overflow")
	}

	vm.Stack[vm.sp] = obj
	vm.sp++
	return nil
}

func (vm *VM) pop() code.Obj {
	o := vm.Stack[vm.sp-1]
	vm.sp--

	return o
}

func (vm *VM) Run() error {
	for ip := 0; ip < len(vm.Instructions); ip++ {
		op := code.Opcode(vm.Instructions[ip])
		switch op {
		case code.OpCreateTable:
			row := vm.Writes[0]
			vm.Writes = vm.Writes[1:]
			decoded := c.DecodeBytes(row[8:])
			tableName := decoded[0]

			offset, err := writeRow(row, TableFile)
			if err != nil {
				return fmt.Errorf("Error writing to table file")
			}

			tNameObj := &code.TableName{Value: tableName}
			err = vm.push(tNameObj)
			if err != nil {
				return fmt.Errorf("Error pushing to stack")
			}
			offObj := &code.RowOffset{Value: offset}
			err = vm.push(offObj)
			if err != nil {
				return fmt.Errorf("Error pushing to stack")
			}
		case code.OpAddTableToTree:
			off := vm.pop()
			col := vm.pop()

			value := off.(*code.RowOffset).Value
			key := col.(*code.TableName).Value

			vm.Pool.Add(key, value)
		case code.OpConstant:
			constIndex := code.ReadUint16(vm.Instructions[ip+1:])
			ip += 2

			err := vm.push(vm.constants[constIndex])
			if err != nil {
				return err
			}
		case code.OpCreateIndex:
			// just pop off the two constants on the stack
			// and update table row
			vm.addIndex(vm.Instructions)
		case code.OpSelect:
		// also need to put a check to see if the where values are an indexed columns
		// get values from stack,
		// need to first find find table row, figure out what index the columns are in table
		// table row: [name, col1, col2, col3, col4]
		// if where clause values are col1 and col3, thos are idxs 1 and 3, loop through all rows
		// skipping all rows not in the correct table, and check those indexes to see if they match
		case code.OpInsert:
			vm.add(vm.Instructions)
		case code.OpUpdate:
			update(vm.Instructions)
		}
	}

	return nil
}

// DQ row, push index to stack
// append ltable row to tables file
// push table offset to stack

// look at next instructions and dq both things in the stack to add to btree
func (vm *VM) addTable(ins []byte) {
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

// use where clause to identify index
// search using index
func (vm *VM) search(ins []byte) []byte {
	vals := c.AccessSelectValues(ins)

	// tableName := vals[0]
	// col := vals[1]
	val := vals[2]
	fmt.Println("len: ", len(val))

	offset, ok := vm.Pool.Search(val)
	if !ok {
		fmt.Println("key not found")
		return nil
	} else {
		row, err := readRow(int64(offset), RowsFile)
		fmt.Println("row found: ", row)
		if err != nil {
			fmt.Println("error: ", err)
		}

		return nil
	}
}

func (vm *VM) add(ins []byte) {

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
