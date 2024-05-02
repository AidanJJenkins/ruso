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
	vm := vm.New()

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

		obj := compile.Compile(program)

		err := vm.Run(obj)
		if err != nil {
			fmt.Println("error running")
		}

		fmt.Println(">>> Executed.")
		// for tok := l.NextToken(); tok.Type != token.EOF; tok = l.NextToken() {
		// 	fmt.Printf("%+v\n", tok)
		// }
	}
}
