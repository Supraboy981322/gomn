package main

import (
	"os"
	"fmt"
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
				file = args[i+1]
				taken = append(taken, i+1)
			 default:
				path = append(path, arg)
			}
		}
	}
}

func main() {
	if file == "" {
		fmt.Fprintln(os.Stderr, "no file provided")
		os.Exit(1)
	}

	GOMN, err := gomn.ParseFile(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse file:  %v", err)
		os.Exit(1)
	}

	var res any
	cur := GOMN
	for _, p := range path {
		if n, ok := cur[p].(gomn.Map); ok {
			cur = n
		} else { res = cur[p] }
	}

	fmt.Print(res)
}
