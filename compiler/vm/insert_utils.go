package vm

import (
	// "encoding/binary"
	"encoding/binary"
	"fmt"
	"slices"

	"github.com/aidanjjenkins/compiler/code"
)

func (vm *VM) createTableObj(name string) *code.TableInfo {
	tObj := &code.TableInfo{Name: name}
	offset, ok := vm.Pool.Search(name)
	decoded := []string{}
	if !ok {
		fmt.Println("Table not found")
		return nil
	} else {
		row, err := readRow(int64(offset), TableFile)
		if err != nil {
			fmt.Println("Error finding table: ", err)
		}

		decoded = DecodeBytes(row)
	}

	tObj.Cols = getColInfo(decoded[1:])
	tObj.Marker = make([]int, len(tObj.Cols))
	nullArr := createNullArrays(len(tObj.Cols))
	tObj.Write = nullArr
	tObj.ColCounter = 0
	tObj.ValCounter = 0
	return tObj
}

func getColInfo(cols []string) []*code.ColCell {
	res := []*code.ColCell{}
	i := 0
	j := 5

	for j <= len(cols) {
		col := cols[i:j]
		newCell := &code.ColCell{}
		newCell.Name = col[0]
		newCell.ColType = col[1]
		if col[2] == "false" {
			newCell.Index = false
		} else {
			newCell.Index = true
		}
		if col[3] == "false" {
			newCell.Unique = false
		} else {
			newCell.Unique = true
		}
		if col[4] == "false" {
			newCell.Pk = false
		} else {
			newCell.Pk = true
		}

		res = append(res, newCell)
		i += 5
		j += 5
	}

	return res
}

func (vm *VM) colCheck(col string, table *code.TableInfo) {
	for i := range table.Cols {
		if table.Cols[i].Name == col {
			table.ColCounter++
			table.Marker[i] = table.ColCounter
			break
		}
	}
}

// this should take an literal obj instead of a string eventually
// that way you can do type checking to make sure the value matches up with the col info
// you can use the table marker to index the table info and check colcells coltype
func (vm *VM) insertVals(value string, table *code.TableInfo) {
	table.ValCounter++
	if table.ColCounter == 0 {
		encoded := encodeString(value)
		table.Write[table.ValCounter-1] = encoded
	} else {
		idx := slices.Index(table.Marker, table.ValCounter)

		encoded := encodeString(value)
		table.Write[idx] = encoded
	}
}

func createNullArrays(size int) [][]byte {
	result := make([][]byte, size)
	for i := 0; i < size; i++ {
		result[i] = []byte{0xFE}
	}
	return result
}

// should look through the table object to see if any of the cols are
// can eventually do the same for unique and non null
func (vm *VM) write(table *code.TableInfo) {
	toWrite := []byte{}

	tName := encodeString(table.Name)
	toWrite = append(toWrite, tName...)

	for i := range table.Write {
		toWrite = append(toWrite, table.Write[i]...)
	}

	l := len(toWrite)

	// do i need this big of a row length?
	lenBuf := make([]byte, RowLen)
	binary.LittleEndian.PutUint32(lenBuf, uint32(l))
	toWrite = append(lenBuf, toWrite...)

	writeRowWithoutIndex(toWrite, RowsFile)

	err := vm.incrememntRowCount(table.Name)
	if err != nil {
		fmt.Println("error: ", err)
		return
	}
}
