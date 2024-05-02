package main

import (
	"fmt"
	"github.com/aidanjjenkins/compiler/repl"
	"os"
)

func assert(cond bool) {
	if !cond {
		panic("assertion failure")
	}
}

func main() {
	// pool := newPool()
	// fmt.Println("Database opened")
	//
	// fmt.Println(pool.db.Path)
	// pool.add("winnie", "stella")
	// pool.search("winnie")
	// fmt.Println("Database closing")
	// user, err := user.Current()
	// if err != nil {
	// 	panic(err)
	// }
	fmt.Println("RusoDB started!")
	fmt.Printf("Feel free to type in commands\n")
	repl.Start(os.Stdin, os.Stdout)
}
