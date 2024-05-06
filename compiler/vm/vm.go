package vm

import (
	"encoding/binary"
	"fmt"

	tree "github.com/aidanjjenkins/bplustree"
	"github.com/aidanjjenkins/compiler/code"
	"github.com/aidanjjenkins/compiler/compile"
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
	Pool         *Pool
	Instructions code.Instructions
}

func New(bytecode *compile.Bytecode) *VM {
	Pool := newPool()
	vm := VM{Pool: Pool, Instructions: bytecode.Instructions}

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
	ins := vm.Instructions
	fmt.Println("instructions: ", ins)
	return nil
}
