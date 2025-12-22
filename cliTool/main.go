package main

import (
	"os"
	"fmt"
	"bufio"
	"slices"
	"github.com/Supraboy981322/gomn"
)

var (
	file string
	path []string
	args = os.Args[1:]
)

func init() {
	var taken []int
	for i, arg := range args {
		if !slices.Contains(taken, i) {
			switch arg {
			 case "-f":
				if len(args) <= i+1 {
					err := fmt.Errorf("called %s arg, but no value provided", arg)
					erorF("invalid arg", err)
				}
				file = args[i+1]
				taken = append(taken, i+1)
			 default:
				path = append(path, arg)
			}
		}
	}
}

func main() {
	var in string
	if file == "" {
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			in += scanner.Text()
		} else if err := scanner.Err(); err != nil {
				erorF("failed read stdin:  %v", err)
		} else {
			err := fmt.Errorf("no input provided")
			erorF("err reading input", err)
		}
	} else {
		inB, err := os.ReadFile(file)
		if err != nil { erorF("failed to read file:  %v", err) }
		in = string(inB)
	}

	GOMN, err := gomn.Parse(in)
	if err != nil { erorF("failed to parse file:  %v", err) }

	var res any
	if len(path) > 0 { 
		cur := GOMN
		for _, p := range path {
			if n, ok := cur[p].(gomn.Map); ok {
					cur = n
			} else { res = cur[p] }
		}
	} else {
	
	}
	fmt.Print(res)
}

func erorF(str string, err error) {
	fmt.Printf("\033[2K")
	str = fmt.Sprintf("\033[1;30;41m%s\033[0m", str)
	err = fmt.Errorf("    \033[1;31m%v\033[0m", err)
	fmt.Fprintf(os.Stderr, "%s\n%v\n", str, err)
	os.Exit(1)
}
