package repl

import (
	"bufio"
	"fmt"
	"io"

	"github.com/aidanjjenkins/compiler/compile"
	"github.com/aidanjjenkins/compiler/lexer"
	"github.com/aidanjjenkins/compiler/parser"
	"github.com/aidanjjenkins/compiler/vm"
)

const PROMPT = ">>> "

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)

	// vm.Pool.Search("dogs")
	for {
		fmt.Printf(PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		l := lexer.New(line)
		p := parser.New(l)

		program := p.ParseProgram()
		if len(p.Errs()) != 0 {
			fmt.Println("error parsing")
			continue
		}

		comp := compile.New()
		err := comp.Compile(program)
		if err != nil {
			fmt.Println("Compile error: ", err)
			return
		}

		machine := vm.New(comp.Bytecode())
		err = machine.Run()
		if err != nil {
			fmt.Println("Error running")
			return
		}

		fmt.Println(">>> Executed.")
		// for tok := l.NextToken(); tok.Type != token.EOF; tok = l.NextToken() {
		// 	fmt.Printf("%+v\n", tok)
		// }
	}
}
