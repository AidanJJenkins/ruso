package vm

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"slices"

	tree "github.com/aidanjjenkins/bplustree"
	"github.com/aidanjjenkins/compiler/code"
	c "github.com/aidanjjenkins/compiler/compile"
)

// ------------------------------------------------
// table row layout
//  len       cell cols...     # of rows in table
// | 8 bytes | unknown amount |      8 bytes      |
// ------------------------------------------------
// cell col layout
// name | type | index | unique | primary key |
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
	constants    []code.Obj
	Stack        []code.Obj
	sp           int
}

func New(bytecode *c.Bytecode) *VM {
	vm := VM{
		Pool:         newPool(),
		Instructions: bytecode.Instructions,
		constants:    bytecode.Constants,
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
		case code.OpConstant:
			opRead := code.ReadUint16(vm.Instructions[ip+1:])
			con := vm.constants[opRead]
			err := vm.push(con)
			if err != nil {
				return err
			}
		case code.OpEncodeStringVal:
			opRead := code.ReadUint16(vm.Instructions[ip+1:])
			encoded := code.EncodedVal{Val: encode(vm.constants[opRead])}
			err := vm.push(&encoded)
			if err != nil {
				return err
			}
		case code.OpEncodeTableCell:
			opRead := code.ReadUint16(vm.Instructions[ip+1:])

			encoded := code.EncodedVal{Val: encode(vm.constants[opRead])}
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
			if t, ok := table.(*code.TableName); ok {
				tName := code.TableName{Value: t.Value}
				err := vm.push(&tName)
				if err != nil {
					return err
				}
			}
		case code.OpWhereCondition:
			opRead := code.ReadUint16(vm.Instructions[ip+1:])
			where := vm.constants[opRead]
			if w, ok := where.(*code.Where); ok {
				col := code.Where{Column: w.Column, Value: w.Value}
				err := vm.push(&col)
				if err != nil {
					return err
				}
			}
		case code.OpCreateTableIndex:
			opRead := code.ReadUint16(vm.Instructions[ip+1:])
			table := vm.constants[opRead]
			switch table := table.(type) {
			case *code.TableName:
				vm.executeAddIndex(table.Value)
			}
		case code.OpSelect:
			numVals := code.ReadUint8(vm.Instructions[ip+1:])
			vm.executeRowSearch(int(numVals))
			for vm.sp > 0 {
				val := vm.pop()
				if v, ok := val.(*code.FoundRow); ok {
					fmt.Printf(">>> %s \n", v.Val)
				}
			}
			// create table object, get table and adds its columns to an array, leave rest null, push to stack
			// as you get check col instructions, mark an array with positions corresponding to what order was given
			// if a table has cols: name, age, breed, weight
			// and the given cols are: age, name, weight
			// mark an array as [0, 0, 0, 0]
			//[0, 1, 0, 0]
			//[2, 1, 0, 0]
			//[2, 1, 0, 3]
			// as you get the values to write, add them to an array that gets pushed to the stack,
			// they should be encoded to bytes, then added to an an array of an array of bytes
			//[7, 0xFE, 0xFE, 0xFE]
			//[7, stella, 0xFE, 0xFE]
			//[7, stella, 0xFE ,20]
		case code.OpInsertRow:
			numVals := code.ReadUint8(vm.Instructions[ip+1:])
			vm.executeRowWrite(int(numVals))
		case code.OpTableInfo:
			opRead := code.ReadUint16(vm.Instructions[ip+1:])
			table := vm.constants[opRead]
			switch table := table.(type) {
			case *code.TableName:
				tObj := vm.createTableObj(table.Value)
				err := vm.push(tObj)
				if err != nil {
					return err
				}
			}
		case code.OpColInfo:
			opRead := code.ReadUint16(vm.Instructions[ip+1:])
			col := vm.constants[opRead]
			switch col := col.(type) {
			case *code.Col:
				tableObj := vm.pop()
				if t, ok := tableObj.(*code.TableInfo); ok {
					vm.colCheck(col.Value, t)
				}
				err := vm.push(tableObj)
				if err != nil {
					return err
				}
			}
		case code.OpValInfo:
			opRead := code.ReadUint16(vm.Instructions[ip+1:])
			col := vm.constants[opRead]
			switch col := col.(type) {
			case *code.Col:
				tableObj := vm.pop()
				if t, ok := tableObj.(*code.TableInfo); ok {
					vm.insertVals(col.Value, t)
				}
				err := vm.push(tableObj)
				if err != nil {
					return err
				}
			}
		case code.OpInsert:
			tableObj := vm.pop()
			if t, ok := tableObj.(*code.TableInfo); ok {
				vm.write(t)
			}
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

func encode(obj code.Obj) []byte {
	switch obj := obj.(type) {
	case *code.TableName:
		return encodeString(obj.Value)
	case *code.ColCell:
		cell := encodeString(obj.Name)
		cell = append(cell, encodeString(obj.ColType)...)
		cell = append(cell, encodeBool(obj.Index))
		cell = append(cell, encodeBool(obj.Unique))
		cell = append(cell, encodeBool(obj.Pk))
		return cell
	case *code.Col:
		return encodeString(obj.Value)
	}
	return nil
}

func encodeString(s string) []byte {
	encodedBytes := []byte(s)
	encodedBytes = append(encodedBytes, 0x00)
	return encodedBytes
}

func encodeBool(b bool) byte {
	if b {
		return 0xFF
	}
	return 0xFD
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
		} else if data[i] == 0xFE {
			result = append(result, " ")
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

// should be nicer way to insert tablename into index probably
func (vm *VM) executeTableWrite(numVals int) error {
	write := []byte{}
	for numVals > 0 {
		val := vm.pop()
		if v, ok := val.(*code.EncodedVal); ok {
			write = append(v.Val, write...)
		}

		numVals -= 1
	}

	count := make([]byte, RowLen)
	binary.LittleEndian.PutUint32(count, uint32(0))
	write = append(write, count...)

	tName := getTableName(write)

	l := len(write)
	lenBuf := make([]byte, RowLen)
	binary.LittleEndian.PutUint32(lenBuf, uint32(l))
	write = append(lenBuf, write...)

	offset, err := writeRow(write, TableFile)
	if err != nil {
		return fmt.Errorf("Error writing table to disk")
	}

	vm.Pool.Add(tName, offset)
	return nil
}

func (vm *VM) incrememntRowCount(tName string) error {
	offset, ok := vm.Pool.Search(tName)
	if !ok {
		return fmt.Errorf("Table not found")
	} else {
		row, err := readRow(int64(offset), TableFile)
		if err != nil {
			fmt.Println("Error finding table: ", err)
		}

		count := DecodeTableCount(row)
		count++

		updateCount := make([]byte, RowLen)
		binary.LittleEndian.PutUint32(updateCount, uint32(count))

		row = row[:len(row)-RowLen]
		row = append(row, updateCount...)

		file, err := os.OpenFile(TableFile, os.O_RDWR, 0644)
		if err != nil {
			fmt.Printf("Failed to open file: %v", err)
		}
		defer file.Close()

		_, err = file.WriteAt(row, int64(offset)+RowLen)
		if err != nil {
			fmt.Printf("Failed to open file: %v", err)
		}

	}

	return nil
}

func getTableName(data []byte) string {
	for i, b := range data {
		if b == 0 {
			return string(data[:i])
		}
	}
	return string(data)
}

func (vm *VM) executeRowWrite(numVals int) {
	write := []byte{}
	for numVals > 0 {
		val := vm.pop()
		if v, ok := val.(*code.EncodedVal); ok {
			write = append(v.Val, write...)
		}

		numVals -= 1
	}

	tName := getTableName(write)
	err := vm.incrememntRowCount(tName)
	if err != nil {
		fmt.Println("error: ", err)
		return
	}

	l := len(write)
	// do i need this big of a row length?
	lenBuf := make([]byte, RowLen)
	binary.LittleEndian.PutUint32(lenBuf, uint32(l))
	write = append(lenBuf, write...)
	writeRowWithoutIndex(write, RowsFile)
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

func (vm *VM) executeRowSearch(numVals int) {
	table := ""
	cols2Find := []string{}
	vals2Find := []string{}

	for numVals > 0 {
		popped := vm.pop()
		switch obj := popped.(type) {
		case *code.TableName:
			table = obj.Value
		case *code.Where:
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

		f := &code.FoundRow{Val: decoded[1:]}
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

func (vm *VM) executeAddIndex(tName string) {
	table := vm.FindTable(tName)
	count := DecodeTableCount(table)
	cols := []string{}
	for vm.sp > 0 {
		val := vm.pop()
		switch v := val.(type) {
		case *code.Col:
			cols = append(cols, v.Value)
		}
	}

	vm.markAsIndex(tName, cols)
	decodedTable := DecodeBytes(table)
	coldIdxs := getColIdxFromTable(decodedTable, cols)

	if count > 0 {
		vm.addExistingRowsToIndex(tName, coldIdxs, count)
	}
}

// using the get coldidxfrom table to find out which item to insert,
// at every row, add length to the "offset" var to keep track of offset,
// if table name is correct, access idx from coldixfrom table that matchs, insert using offset and col val
func (vm *VM) addExistingRowsToIndex(tName string, colIdx []int, count int) error {
	// offset := 0
	file, err := os.Open(RowsFile)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return nil
	}
	defer file.Close()

	rowsChecked := 0
	for {

		if rowsChecked >= count {
			fmt.Println("checked all rows in table")
			break
		}
		offset, err := file.Seek(0, io.SeekCurrent)
		if err != nil {
			fmt.Printf("Error getting offset: %v\n", err)
			return nil
		}

		lengthBytes := make([]byte, 8)
		_, err = file.Read(lengthBytes)
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

		if decoded[0] == tName {
			for i := range colIdx {
				// fmt.Println("col val: ", decoded[colIdx[i]])
				// fmt.Println("offset: ", offset)

				vm.Pool.Add(decoded[colIdx[i]], uint32(offset))
			}

		}
	}

	return nil
}

func (vm *VM) markAsIndex(tName string, cols []string) {
	offset, ok := vm.Pool.Search(tName)
	decoded := []string{}
	if !ok {
		fmt.Println("Table not found")
		return
	} else {
		row, err := readRow(int64(offset), TableFile)
		if err != nil {
			fmt.Println("Error finding table: ", err)
		}

		decoded = DecodeBytes(row)
	}

	for i := range cols {
		idx := slices.Index(decoded, cols[i])
		decoded[idx+2] = "true"
	}

	buf := []byte{}
	for i := range decoded {
		if decoded[i] == "false" {
			b := encodeBool(false)
			buf = append(buf, b)
		} else if decoded[i] == "true" {
			b := encodeBool(true)
			buf = append(buf, b)
		} else {
			encoded := encodeString(decoded[i])
			buf = append(buf, encoded...)
		}
	}

	file, err := os.OpenFile(TableFile, os.O_RDWR, 0644)
	if err != nil {
		fmt.Println("Error opening table file: ", err)
	}
	defer file.Close()

	_, err = file.WriteAt(buf, int64(offset)+int64(RowLen))
	if err != nil {
		fmt.Println("Error writing table file: ", err)
	}
}
