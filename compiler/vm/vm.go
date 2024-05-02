package vm

import (
	"encoding/binary"
	"fmt"
	"os"

	tree "github.com/aidanjjenkins/bplustree"
	// "github.com/aidanjjenkins/compiler/code"
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
	// assert(err == nil)
	return p
}

type VM struct {
	Pool *Pool
}

func New() *VM {
	Pool := newPool()
	vm := VM{Pool: Pool}

	return &vm
}

func (p *Pool) add(key []byte, val uint32) {
	offsetBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(offsetBytes, val)
	p.db.Set(key, offsetBytes)
	p.Search(string(key))
	// p.ref[key] = val
}

func (p *Pool) Search(key string) {
	bytes, found := p.db.Get(key)
	if found {
		fmt.Println(">>> ", bytes)
		value := binary.LittleEndian.Uint32(bytes)
		fmt.Println("value:", value)
	} else {
		fmt.Println(">>> Value not found")
	}
}

func (vm *VM) Run() error {
	return nil
}

func writeToRowsFile(value []byte) (uint32, error) {
	// Open or create a file to write the value
	file, err := os.OpenFile("rows.db", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	// Get the current offset in the file
	fileInfo, err := file.Stat()
	if err != nil {
		return 0, err
	}
	offset := uint32(fileInfo.Size())
	fmt.Println("offset: ", offset)

	// Write the value to the file
	_, err = file.Write(value)
	if err != nil {
		return 0, err
	}

	// Return the offset (location) and length of the row in the file
	return offset, nil
}

// func accessMetaData() map[string]any {
// 	m := make(map[string]any)
// }

func readFromRows(offset int64) ([]byte, error) {
	// Open the file to read from
	file, err := os.Open("rows.db")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Seek to the specified offset
	_, err = file.Seek(offset, 0)
	if err != nil {
		return nil, err
	}

	// Read the length of the value
	lengthBytes := make([]byte, 4) // Assuming int64 for length
	_, err = file.Read(lengthBytes)
	if err != nil {
		return nil, err
	}
	length := int64(binary.LittleEndian.Uint32(lengthBytes))

	// Read the value from the file
	value := make([]byte, length)
	_, err = file.Read(value)
	if err != nil {
		return nil, err
	}

	return value, nil
}
