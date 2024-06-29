package vm

import (
	"encoding/binary"
	"fmt"
	"os"
	"slices"

	tree "github.com/aidanjjenkins/bplustree"
	"github.com/aidanjjenkins/compiler/code"
	c "github.com/aidanjjenkins/compiler/compile"
	o "github.com/aidanjjenkins/compiler/object"
)

// ------------------------------------------------
// TABLE LAYOUT
//  len       cell cols...     # of rows in table
// | 8 bytes | unknown amount |      8 bytes      |
// ------------------------------------------------
// cell col layout
// name | type | index | unique | primary key |
// ------------------------------------------------

// ------------------------------------------------
//  ROW LAYOUT
//  len       values
// | 8 bytes | unknown amount|
// ------------------------------------------------
// values string layout
// value | null byte|
// ------------------------------------------------

const TableFile string = "tables.db"
const IdxFile string = "index.db"
const RowsFile string = "rows.db"
const StackSize = 2040
const RowLen = 8

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
	constants    []o.Obj
	Stack        []o.Obj
	sp           int
}

func New(bytecode *c.Bytecode) *VM {
	vm := VM{
		Pool:         newPool(),
		Instructions: bytecode.Instructions,
		constants:    bytecode.Constants,
		Stack:        make([]o.Obj, StackSize),
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

func (vm *VM) push(obj o.Obj) error {
	if vm.sp >= StackSize {
		return fmt.Errorf("Stack overflow")
	}

	vm.Stack[vm.sp] = obj
	vm.sp++
	return nil
}

func (vm *VM) pop() o.Obj {
	o := vm.Stack[vm.sp-1]
	vm.sp--

	return o
}

// make sure index still works, i chagned the opcode from Opconstant to OpAddIndex
func (vm *VM) Run() error {
	for ip := 0; ip < len(vm.Instructions); ip++ {
		op := code.Opcode(vm.Instructions[ip])
		switch op {
		case code.OpConstant:
			opRead := code.ReadUint16(vm.Instructions[ip+1:])
			con := vm.constants[opRead]
			err := vm.push(con)
			if err != nil {
				return err
			}
		case code.OpEncodeStringVal:
			opRead := code.ReadUint16(vm.Instructions[ip+1:])
			encoded := o.EncodedVal{Val: encode(vm.constants[opRead])}
			err := vm.push(&encoded)
			if err != nil {
				return err
			}
		case code.OpEncodeTableCell:
			opRead := code.ReadUint16(vm.Instructions[ip+1:])

			encoded := o.EncodedVal{Val: encode(vm.constants[opRead])}
			err := vm.push(&encoded)
			if err != nil {
				return err
			}
		case code.OpCreateTable:
			numVals := code.ReadUint8(vm.Instructions[ip+1:])
			vm.executeTableWrite(int(numVals))
		case code.OpTableNameSearch:
			opRead := code.ReadUint16(vm.Instructions[ip+1:])
			table := vm.constants[opRead]
			if t, ok := table.(*o.TableName); ok {
				tName := o.TableName{Value: t.Value}
				err := vm.push(&tName)
				if err != nil {
					return err
				}
			}
		case code.OpWhereCondition:
			opRead := code.ReadUint16(vm.Instructions[ip+1:])
			where := vm.constants[opRead]
			if w, ok := where.(*o.Where); ok {
				col := o.Where{Column: w.Column, Value: w.Value}
				err := vm.push(&col)
				if err != nil {
					return err
				}
			}
		case code.OpCreateTableIndex:
			opRead := code.ReadUint16(vm.Instructions[ip+1:])
			table := vm.constants[opRead]
			switch table := table.(type) {
			case *o.TableName:
				vm.executeAddIndex(table.Value)
			}
		case code.OpSelect:
			numVals := code.ReadUint8(vm.Instructions[ip+1:])
			vm.executeRowSearch(int(numVals))
			for vm.sp > 0 {
				val := vm.pop()
				if v, ok := val.(*o.FoundRow); ok {
					fmt.Printf(">>> %s \n", v.Val)
				}
			}

		// create a table object then decide what to do
		case code.OpCheckCol:
			// begin refactor tomorrow, need cases for colcheck, values and mabye table?
			// if refactor is quick, figure out null values?
		case code.OpInsertRow:
			numVals := code.ReadUint8(vm.Instructions[ip+1:])
			vm.executeRowWrite(int(numVals))
		case code.OpUpdate:
		}
	}

	return nil
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

func readRow(offset int64, filename string) ([]byte, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

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

func DecodeBytes(data []byte) []string {
	var result []string

	for i := 0; i < len(data); i++ {
		if data[i] == 0x00 {
			break
		} else if data[i] == 0xFF {
			result = append(result, "true")
		} else if data[i] == 0xFD {
			result = append(result, "false")
		} else {
			str := ""
			for j := i; j < len(data) && data[j] != 0x00; j++ {
				str += string(data[j])
				i++
			}
			result = append(result, str)
		}
	}

	return result
}

func DecodeTableCount(data []byte) int {
	counter := binary.LittleEndian.Uint32(data[len(data)-8:])
	return int(counter)
}

func getTableName(data []byte) string {
	for i, b := range data {
		if b == 0 {
			return string(data[:i])
		}
	}
	return string(data)
}

func (vm *VM) executeRowSearch(numVals int) {
	table := ""
	cols2Find := []string{}
	vals2Find := []string{}

	for numVals > 0 {
		popped := vm.pop()
		switch obj := popped.(type) {
		case *o.TableName:
			table = obj.Value
		case *o.Where:
			cols2Find = append([]string{obj.Column}, cols2Find...)
			vals2Find = append([]string{obj.Value}, vals2Find...)
		}
		numVals -= 1
	}

	tableBytes := vm.FindTable(table)
	decoded := DecodeBytes(tableBytes)
	tName := decoded[0]
	decoded = decoded[1:]

	tablesCols := []string{}
	i := 0
	for i < len(decoded) {
		tablesCols = append(tablesCols, decoded[i])
		i += 5
	}

	idxs := getColIdxFromTable(tablesCols, cols2Find)

	vm.walkTable(RowsFile, tName, idxs, vals2Find)
}

func getColIdxFromTable(table, cols []string) []int {
	idxs := []int{}

	for i := range cols {
		idx := slices.Index(table, cols[i])
		if idx != -1 {
			idxs = append(idxs, idx)
		}
	}

	return idxs
}

// full table scan
func (vm *VM) walkTable(filePath, tName string, idxs []int, vals []string) []string {
	file, err := os.Open(filePath)
	found := []string{}
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return nil
	}
	defer file.Close()

	for {
		lengthBytes := make([]byte, 8)
		_, err := file.Read(lengthBytes)
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			fmt.Printf("Error reading length: %v\n", err)
			return nil
		}

		rowLength := binary.LittleEndian.Uint32(lengthBytes)

		rowData := make([]byte, rowLength)
		_, err = file.Read(rowData)
		if err != nil {
			fmt.Printf("Error reading row: %v\n", err)
			return nil
		}

		decoded := DecodeBytes(rowData)
		if decoded[0] != tName {
			continue
		}
		eq := searchRow(idxs, decoded[1:], vals)
		if !eq {
			continue
		}

		f := &o.FoundRow{Val: decoded[1:]}
		vm.push(f)
	}

	return found
}

func searchRow(idxs []int, row, vals []string) bool {
	check := []string{}
	for _, i := range idxs {
		check = append(check, row[i])
	}

	eq := slices.Equal(check, vals)
	if !eq {
		return false
	}

	return true
}
