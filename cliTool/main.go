package main

import (
	"os"
	"fmt"
	"bufio"
	"slices"
	"strings"
	"strconv"
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
				erorF("failed read stdin", err)
		} else {
			err := fmt.Errorf("no input provided")
			erorF("err reading input", err)
		}
	} else {
		inB, err := os.ReadFile(file)
		if err != nil { erorF("failed to read file", err) }
		in = string(inB)
	}

	GOMN, err := gomn.Parse(in)
	if err != nil {
		erorF("failed to parse gomn", fmt.Errorf("\r%v", err))
	}

	if len(path) > 0 { 
		var res any
		cur := GOMN
		for _, p := range path {
			if n, ok := cur[p].(gomn.Map); ok {
					cur = n
			} else { res = cur[p] }
		}
		fmt.Print(res)
	} else {
		var typ string
		isKey := true
		var esc bool
		li, col := 1, 0
		for _, c := range in {
			var cur string
			col++
			if typ == "num" { typ = "" }
			_, err := strconv.Atoi(string(c))
			if err == nil && typ == "" { typ = "num" }
			switch c {
			 case '/':
				if !esc && (typ == "" || typ == "comment") {
					typ = "comment"
					cur += "\033[0;90m"+string(c)+"\033[0m"
				} else if !esc && (typ == "multi-comment") {
					typ = ""
					cur += "\033[0;90m"+string(c)+"\033[0m"
				}
			 case '*':
				if !esc && (typ == "comment") {
					typ = "multi-comment"
					cur += "\033[0;90m"+string(c)+"\033[0m"
				} else if typ == "multi-comment" {
					cur += "\033[0;90m"+string(c)+"\033[0m"
				}
			 case '\\': esc = !esc;
			 case '[':
				if !esc && typ == "" { isKey = true }
				fallthrough
			 case ']', '{', '}', '|', ',':
				if !esc && typ == "" {
					cur += "\033[1;97m"+string(c)+"\033[0m"
				} else if !esc && !strings.Contains(typ, "comment") {
					invGomn(li, col, c, typ)
				} else { cur += "\033[90m"+string(c)+"\033[0m" }
       case ':', '=':
				if !esc && typ == "" {
					cur += "\033[36m"+string(c)+"\033[0m"
					isKey = false
				} else if !esc && !strings.Contains(typ, "comment") {
					invGomn(li, col, c, typ)
				} else { cur += "\033[90m"+string(c)+"\033[0m" }
			 case 'f', 't':
				if typ == "" && !esc { typ = "bool" }
				fallthrough
			 case 'a', 'l', 's', 'e', 'r', 'u': fallthrough
			 case '"':
				if c == '"' && !strings.Contains(typ, "comment") {
					if !esc && typ == "str" { typ = ""
					} else if !esc && typ == "" { typ = "str" }
				} else if !esc && strings.Contains(typ, "comment") {
					cur += "\033[90m"+string(c)+"\033[0m" 
					fmt.Print(cur)
					continue
				};fallthrough
			 case '\'':
				if c == '\'' {
					if !esc && typ == "rune" { typ = ""
					} else if !esc { typ = "rune" }
				}; fallthrough
			 case '`':
				if c == '`' {
					if !esc && typ == "byte" { typ = ""
					} else if !esc { typ = "byte" }
				}; fallthrough
			 default:
				if isKey { cur += "\033[0;1m"
				} else { cur += "\033[0m" }
				if typ == "str" || (!esc && c == '"') {
					if isKey { cur += "\033[94m"
					} else { cur += "\033[32m" }
				} else if typ == "num" || typ == "bool" {
					cur += "\033[38;2;255;165;0m"
					if c == 'e' { typ = "" }
				} else if typ == "comment" || typ == "multi-comment" {
					cur += "\033[3;90m"
					if c == '\n' && typ == "comment" { typ = "" }
				} else if typ == "rune" || (!esc && c == '\'') {
					cur += "\033[95m"
				} else if typ == "byte" || (!esc && c == '`') {
					cur += "\033[33m"
				} else if typ == "" {
					switch c {
					 case '\\', ' ', '\n', '\'': func(){}()
					 default: invGomn(li, col, c, "none")
					}
					cur += "\033[0m"
				}
				cur += string(c)+"\033[0m"
				if esc { esc = false }
				if c == '\n' { li++ ; col = 0 }
			}

			fmt.Print(cur)
		}
	}
}

func erorF(str string, err error) {
	fmt.Printf("\033[2K")
	str = fmt.Sprintf("\033[1;30;41m%s\033[0m", str)
	err = fmt.Errorf("    \033[1;31m%v\033[0m", err)
	fmt.Fprintf(os.Stderr, "%s\n%v\n", str, err)
	os.Exit(1)
}

func invGomn(li int, col int, c rune, typ string) {
	fmt.Printf("\033[1;5;7;41m%s\033[0m", string(c))
	err := fmt.Errorf("parsing as type '\033[97m%s\033[31m',"+
										"but got '\033[97m%s\033[31m'", typ, string(c))
	str := fmt.Sprintf("invalid gomn (line %d col %d)", li, col)
	fmt.Println("\n")
	erorF(str, err)
}
