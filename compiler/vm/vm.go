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

type Pool struct {
	db tree.Pager
	// ref map[string]string
}

func newPool() *Pool {
	p := &Pool{}
	// p.ref = map[string]string{}
	p.db.Path = "db.db"
	err := p.db.Open()
	if err != nil {
		fmt.Println("err: ", err)
	}
	return p
}

type VM struct {
	Pool         *Pool
	Instructions code.Instructions
}

func New(bytecode *c.Bytecode) *VM {
	vm := VM{Pool: newPool(), Instructions: bytecode.Instructions}

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

func (vm *VM) Run() error {
	ins := vm.Instructions

	op := code.Opcode(c.AccessFirstByte(ins))
	switch op {
	case code.OpCreateTable:
		vm.addTable(vm.Instructions)
	case code.OpCreateIndex:
		addIndex(vm.Instructions)
	case code.OpSelect:
		search(vm.Instructions)
	case code.OpInsert:
		add(vm.Instructions)
	case code.OpUpdate:
		update(vm.Instructions)
	}
	return nil
}

func (vm *VM) addTable(ins []byte) {
	offset, err := writeRow(ins[1:], "test.db")
	if err != nil {
		fmt.Println("error: ", err)
	}

	name := accessTableNameBytes(ins[1:])

	vm.Pool.Add([]byte(name), offset)
}

func (vm *VM) FindTable(name string) []byte {
	offset, ok := vm.Pool.Search(name)
	if !ok {
		fmt.Println("key not found")
		return nil
	} else {
		row, err := readRow(int64(offset))
		if err != nil {
			fmt.Println("error: ", err)
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

	n, err := file.Write(data)
	if err != nil {
		return 0, err
	}

	offset := uint32(n) - uint32(len(data))
	return offset, nil
}

func readRow(offset int64) ([]byte, error) {
	file, err := os.Open("test.db")
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

// func accessTableMetaDataIndex(ins []byte) {
// 	fmt.Println("MetaData Index", ins[:1+c.TableIndexSize])
// }

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
		// If no null byte or whitespace character found, endIndex is the end of the substring
		endIndex = len(substring)
	}

	// Trim the substring to the first null byte or whitespace character
	trimmedSubstring := substring[:endIndex]

	return string(trimmedSubstring)
}

// for the insertion
func accessTableNameBytes(ins []byte) string {
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

func addIndex(ins []byte) {
	fmt.Println("need to implement add index")
	fmt.Println("instructions: ", ins)
}

func search(ins []byte) {
	fmt.Println("need to implement search")
	fmt.Println("instructions: ", ins)
}

func add(ins []byte) {
	fmt.Println("need to implement add")
	fmt.Println("instructions: ", ins)
}

func update(ins []byte) {
	fmt.Println("need to implement update")
	fmt.Println("instructions: ", ins)
}
