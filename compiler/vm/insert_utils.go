package vm

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"slices"

	o "github.com/aidanjjenkins/compiler/object"
)

func encode(obj o.Obj) []byte {
	switch obj := obj.(type) {
	case *o.TableName:
		return encodeString(obj.Value)
	case *o.ColCell:
		cell := encodeString(obj.Name)
		cell = append(cell, encodeString(obj.ColType)...)
		cell = append(cell, encodeBool(obj.Index))
		cell = append(cell, encodeBool(obj.Unique))
		cell = append(cell, encodeBool(obj.Pk))
		return cell
	case *o.Col:
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

func (vm *VM) executeRowWrite(numVals int) {
	write := []byte{}
	for numVals > 0 {
		val := vm.pop()
		if v, ok := val.(*o.EncodedVal); ok {
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

func (vm *VM) executeAddIndex(tName string) {
	table := vm.FindTable(tName)
	count := DecodeTableCount(table)
	cols := []string{}
	for vm.sp > 0 {
		val := vm.pop()
		switch v := val.(type) {
		case *o.Col:
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

// should be nicer way to insert tablename into index probably
func (vm *VM) executeTableWrite(numVals int) error {
	write := []byte{}
	for numVals > 0 {
		val := vm.pop()
		if v, ok := val.(*o.EncodedVal); ok {
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
