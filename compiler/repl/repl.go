package repl

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strings"

	c "github.com/aidanjjenkins/compiler/compile"
	"github.com/aidanjjenkins/compiler/lexer"
	"github.com/aidanjjenkins/compiler/parser"
	"github.com/aidanjjenkins/compiler/vm"
)

// can now loop through table rows can get each table from the file
// now need to access the value from the table to create the in memory table info

const PROMPT = ">>> "

func onStart() map[string]*vm.Tbls {
	m := make(map[string]*vm.Tbls)
	fileInfo, err := os.Stat(vm.TableFile)

	if os.IsNotExist(err) {
		return m
	}

	isEmpty := fileInfo.Size() == 0
	if isEmpty {
		return m
	}
	m, err = readTableData()
	if err != nil {
		fmt.Println("Error reading table data", err)
		return nil
	}

	return m
}

func readTableData() (map[string]*vm.Tbls, error) {
	file, err := os.Open(vm.TableFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	fmt.Println("fileinfo", fileInfo.Size())
	if err != nil {
		fmt.Println("Next offset is not withing file size", err)
		return nil, err
	}

	fileSize := fileInfo.Size()
	fileBytes := make([]byte, fileSize)
	file.Read(fileBytes)

	tables := [][]byte{}
	for len(fileBytes) > 0 {
		row, newFileBytes := readTableRow(fileBytes)
		tables = append(tables, row)
		fileBytes = newFileBytes
	}

	m := make(map[string]*vm.Tbls)
	for _, t := range tables {
		name := cleanse(t[c.TableMetaDataSize : c.TableMetaDataSize+c.TableNameSize])
		cols := t[c.TableMetaDataSize+c.TableNameSize:]
		lens := AccessTableMetaDataLengths(t)
		colMap := make(map[string]string)
		for i := 0; i < len(lens); i += 2 {
			valueLength := lens[i]
			typeLength := lens[i+1]

			grouplen := valueLength + typeLength
			group := cols[:grouplen]

			colName := string(group[:valueLength])
			colType := string(group[valueLength:])
			colMap[colName] = colType

			m[name] = &vm.Tbls{Cols: colMap, Idx: []string{}}
			cols = cols[grouplen:]
		}

	}
	// fmt.Println("wishlist: ", m["wishlist"].Cols)
	// fmt.Println("dogs: ", m["dogs"].Cols)

	return m, nil
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

func AccessTableMetaDataLengths(ins []byte) []int {
	lens := ins[c.TableIndexSize : c.TableIndexSize+c.TableMetaDataLengths]
	l := []int{}

	for i := range lens {
		if lens[i] != 0 {
			l = append(l, int(lens[i]))
		}
	}

	return l
}

func readTableRow(file []byte) ([]byte, []byte) {
	totalLengthBuf := file[:8]

	totalLength := binary.LittleEndian.Uint32(totalLengthBuf)
	row := file[8 : totalLength+8]
	file = file[totalLength+8:]
	return row, file
}

func Start(in io.Reader, out io.Writer) {
	dbInfo := onStart()
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

				comp := c.New()
				err := comp.Compile(program)
				if err != nil {
					fmt.Println("Compile error: ", err)
					return
				}

				machine := vm.New(comp.Bytecode(), dbInfo)
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
