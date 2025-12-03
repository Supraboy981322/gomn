package gomn

import (
	"os"
	"fmt"
	"errors"
	"strconv"
	"strings"
	"unicode"
	"encoding/gob"
)

type (
	Map map[interface{}]interface{}

	parser struct {
		s   string
		pos int
		n   int
	}
)

func init() {
	gob.Register(Map{})
}

//parse and ignore err
func ParseIgn(input string) Map {
	res, _ := Parse(input) 
	return res
}

func GetValue(key any, gomn Map) (interface{}, bool) {
	if gomn[key] == nil {
		return nil, false
	}
	return gomn[key], true
}

func GetValueFromStr(key any, gomnStr string) (interface{}, error) {
	gomnMap, err := Parse(gomnStr)
	if err != nil {
		return nil, err
	}

	return gomnMap[key], nil 
}

func ParseFile(file string) (Map, error) {
	fileBytes, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	return Parse(string(fileBytes))
}

//func called from mod
func Parse(input string) (Map, error) {
	//construct parser 
	p := &parser{
		s: input,
		pos: 0,
		n: len(input),
	}

	//make a blank map
	out := make(Map)
	for {
		p.skipSpaces()
		if p.eof() {
			break
		}
 
		//check if valid start
		if !p.consume('[') {
			return nil, p.errf("expected '[' to start key")
		}
 
		key, err := p.parseValueSimple()
		if err != nil {
			return nil, err
		}

		//make sure key is ended properly
		if !p.consume(']') {
			return nil, p.errf("expected ']' after key")
		}
		
		p.skipSpaces()

		//make sure it's defined 
		if !p.consume(':') || !p.consume('=') {
			return nil, p.errf("expected ':=' after key")
		}

		p.skipSpaces()

		//get the value
		val, err := p.parseValue()
		if err != nil {
			return nil, err
		}
 
		//add to result
		out[key] = val

		p.skipSpaces()
	}

	return out, nil
}

//parses for values (obviously)
func (p *parser) parseValue() (interface{}, error) {
	p.skipSpaces()
	if p.eof() {
		return nil, p.errf("unexpected EOF while parsing value")
	}

	c := p.peek()
	switch c {
	case '{':
		return p.parseArray() //array
	case '[':
		return p.parseValueSimple()
	case '"':
		return p.parseString() //returns as string
	case '\'':
		return p.parseString() //made into rune before return 
	case '|':
		return p.parseMap() //nested map
	default:
		return p.parseValueSimple()
	}
}

func (p *parser) parseMap() (Map, error) {
	p.pos++ //fixes an odd bug,
	        //  probably a logic issue

	//make a blank map
	out := make(Map)
	for {
		p.skipSpaces()
		if p.next() == '|' {
			break //assumed end of nested map
		}

		//this fixes said bug from earlier too 
		p.pos--

		p.skipSpaces()
		
		//check if the file has ended
		if p.eof() {
			return out, p.errf("unexpected EOF while parsing nested map")
		}

		p.skipSpaces()

		//check if valid start
		if !p.consume('[') {
			return nil, p.errf("nested: expected '[' to start key")
		}
 
		//
		key, err := p.parseValueSimple()
		if err != nil {
			return nil, err
		}

		//make sure key is ended properly
		if !p.consume(']') {
			return nil, p.errf("nested: expected ']' after key")
		}
		
		p.skipSpaces()

		//make sure it's defined 
		if !p.consume(':') || !p.consume('=') {
			return nil, p.errf("nested: expected ':=' after key")
		}

		p.skipSpaces()

		//get the value
		val, err := p.parseValue()
		if err != nil {
			return nil, err
		}
 
		//add to result
		out[key] = val

		p.skipSpaces()
	}

	return out, nil
}

//parses primitive single literal, nested constructs, or maps
func (p *parser) parseValueSimple() (interface{}, error) {
	p.skipSpaces()
	if p.eof() {
		return nil, p.errf("unexpected EOF in literal")
	}

	//check if file is about to end
	switch p.peek() {
	case '"':
		return p.parseString()
	case '\'':
		return p.parseString() //is returned as a []rune
	case '{':
		return p.parseArray() //assume it's an array
	case '[':
		p.consume('[')

		//parse key
		v, err := p.parseValueSimple()
		if err != nil {
			return nil, err
		}
		
		//make sure the key is ended properly
		if !p.consume(']') {
			return nil, p.errf("expected ']' after key value")
		}

		return v, nil
	default:
		return p.parseIdentifierOrNumber()
	}
}

func (p *parser) parseArray() (interface{}, error) {
	//make sure it's actually an array
	if !p.consume('{') {
		return nil, p.errf("expected '{' to start array")
	}
	
	//create the interface for the array
	arr := make([]interface{}, 0)
	//iterate through chars
	for {
		p.skipSpaces()

		//make sure the array is closed properly 
		if p.eof() {
			return nil, p.errf("unexpected EOF in array")
		}

		//close array, assume no more items
		if p.peek() == '}' {
			p.consume('}')
			break
		}

		//get val of item
		v, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		
		//add item to array
		arr = append(arr, v)

		p.skipSpaces()

		//check if there's another item
		if p.peek() == ',' {
			p.consume(',')
			p.skipSpaces()
			continue
		}

		//check if it's ended
		if p.peek() == '}' {
			p.consume('}')
			break
		}

		//assume something invalid if got this far 
		return nil, p.errf("expected ',' or '}' in array")
	}

	return arr, nil
}

func (p *parser) parseString() (interface{}, error) {
	//make sure its formatting is valid 
	if !p.consume('"') && !p.consume('\'') {
		return nil, p.errf("expected '\"' or '\\'' to start string")
	}
 
	var sb strings.Builder
	
	escaped := false
	for !p.eof() {
		//get next char 
		ch := p.next()

		//check whether to escape it
		if escaped {
			switch ch {
			case 'n':
				sb.WriteByte('\n')
			case 'r':
				sb.WriteByte('\r')
			case 't':
				sb.WriteByte('\t')
			case '\\':
				sb.WriteByte('\\')
			case '"':
				sb.WriteByte('"')
			case '\'':
				sb.WriteByte('\'')
			default:
				sb.WriteByte(ch)
			}

			//return back to unescaped
			escaped = false

			continue
		}

		//escape escape
		if ch == '\\' {
			escaped = true
			continue
		}

		//assume ended 
		if ch == '"' {
			return sb.String(), nil
		}

		//assume ended, also this is for []runes
		if ch == '\'' {
			return []rune(sb.String()), nil
		}
 
		sb.WriteByte(ch)
	}

	return nil, p.errf("unterminated string")
}

func (p *parser) parseIdentifierOrNumber() (interface{}, error) {
	start := p.pos
	if p.peek() == '-' || p.peek() == '+' {
		p.next()
	}
	if unicode.IsDigit(rune(p.peek())) {
		for !p.eof() &&
				(unicode.IsDigit(rune(p.peek())) ||
						 p.peek() == '.') {
			p.next()
		}

		raw := p.s[start:p.pos]
		if strings.Contains(raw, ".") {
			f, err := strconv.ParseFloat(raw, 64)
			if err != nil {
				return nil, p.errf("invalid float %q", raw)
			}

			return f, nil
		}

		i, err := strconv.ParseInt(raw, 10, 64)
		if err == nil {
			return int(i), nil
		}

		f, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return nil, p.errf("invalid number %q", raw)
		}

		return f, nil
	}

	for !p.eof() && 
			(unicode.IsLetter(rune(p.peek())) || 
					 p.peek() == '_' ||
					 unicode.IsDigit(rune(p.peek()))) {
		p.next()
	}

	raw := p.s[start:p.pos]
	switch raw {
	case "true":
		return true, nil
	case "false":
		return false, nil
	case "null", "nil":
		return nil, nil
	default:
		if raw == "" {
			return nil, p.errf("unexpected token: %v\n  at pos %d",
										string(p.s[p.pos]),
										p.pos)
		}

		return raw, nil
	}
}


func (p *parser) skipSpaces() {
	for !p.eof() {
		c := p.peek()

		if unicode.IsSpace(rune(c)) {
			p.next()

			continue
		}

		// C-style comments
		if c == '/' && p.pos+1 < p.n && p.s[p.pos+1] == '/' {
			// consume until newline
			p.pos += 2

			for !p.eof() && p.peek() != '\n' {
				p.next()
			}

			continue
		}

		// basic mult-line comments
		if c == '/' && p.pos+1 < p.n && p.s[p.pos+1] == '*' {
			p.pos += 2

			for !p.eof() {
				if p.peek() == '*' && p.pos+1 < p.n && p.s[p.pos+1] == '/' {
					p.pos += 2

					break
				}

				p.next()
			}

			continue
		}

		break
	}
}

//make sure the the file isn't about to end
func (p *parser) peek() byte {
	if p.eof() {
		return 0
	}

	return p.s[p.pos]
}

//get next char and increment position
func (p *parser) next() byte {
	if p.eof() {
		return 0
	}

	ch := p.s[p.pos]

	p.pos++
	
	return ch
}

func (p *parser) consume(expected byte) bool {
	if p.eof() {
		return false
	}

	//increment position and return valid 
	if p.s[p.pos] == expected {
		p.pos++

		return true
	}

	return false
}

//end of file check
func (p *parser) eof() bool {
	return p.pos >= p.n
}

//error formatter (yell at user consistently)
func (p *parser) errf(format string, args ...interface{}) error {
	msg := fmt.Sprintf(format, args...)

	//get the start position of snippet 
	start := p.pos - 10
	if start < 0 {
		start = 0
	}

	//get the end position of snippet 
	end := p.pos + 10
	if end > p.n {
		end = p.n
	}

	//get text before invalid
	beforeBad := "\033[0;37m" + p.s[start:p.pos] + "\033[0m"

	//highlight invalid char
	badChar := "\033[1;31m" + string(p.s[p.pos]) + "\033[0m"

	//get text after invalid
	afterBad := "\033[0;37m" + p.s[p.pos+1:end] + "\033[0m"

	//build the snippet
	snippet := beforeBad + badChar + afterBad

	//create invalid char pointer
	pointer := "    \033[1;31m" //starting point and color code 
	for i := 0; i < len(p.s[start:p.pos]); i++ {
		pointer += " " //add whitespace
	}; pointer += "^\033[0m" //add pointer and end color code

	//construct err
	errStr := fmt.Sprintf("%s near\n    %s\n%s", 
			msg, snippet, pointer)

	//return err
	return errors.New(errStr)
}
