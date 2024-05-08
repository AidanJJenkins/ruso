package repl

import (
	"bufio"
	"fmt"
	"github.com/aidanjjenkins/compiler/compile"
	"github.com/aidanjjenkins/compiler/lexer"
	"github.com/aidanjjenkins/compiler/parser"
	"github.com/aidanjjenkins/compiler/vm"
	"io"
	"strings"
)

const PROMPT = ">>> "

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)

	for {
		fmt.Printf(PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		command := scanner.Text()

		if strings.HasPrefix(command, "\\") {
			switch command {
			case "\\q":
				fmt.Println(">>> Shutting down...")
				return
			default:
				fmt.Println(">>> Unknown meta command:", command)
			}
		} else {
			if string(command[len(command)-1]) != ";" {
				fmt.Println("Missing ';'")
			} else {
				l := lexer.New(command)
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
			}
		}

	}
}
