package main

import (
	"fmt"

	tree "github.com/aidanjjenkins/bplustree"
)

func assert(cond bool) {
	if !cond {
		panic("assertion failure")
	}
}

type Pool struct {
	db  tree.Pager
	ref map[string]string
}

func newPool() *Pool {
	p := &Pool{}
	p.ref = map[string]string{}
	p.db.Path = "db.db"
	err := p.db.Open()
	if err != nil {
		fmt.Println("err: ", err)
	}
	// assert(err == nil)
	return p
}

func (p *Pool) add(key string, val string) {
	p.db.Set([]byte(key), []byte(val))
	p.ref[key] = val
}

func (p *Pool) search(key string) {
	bytes, found := p.db.Get(key)
	if found {
		fmt.Println(">>> ", string(bytes))
	} else {
		fmt.Println(">>> Value not found")
	}
}

func main() {
	pool := newPool()
	fmt.Println("Database opened")

	fmt.Println(pool.db.Path)
	// pool.add("winnie", "stella")
	pool.search("winnie")
	fmt.Println("Database closing")
}
