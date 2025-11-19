package gomn

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

type (
	Map map[interface{}]interface{}

	parser struct {
		s   string
		pos int
		n   int
	}
)

//parse and ignore err
func ParseIgn(input string) Map {
	res, _ := Parse(input) 
	return res
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
	p.pos++
	//make a blank map
	out := make(Map)
	for {
		p.skipSpaces()
		if p.next() == '|' {
			break //assumed end of nested map
		}
		p.pos--

		p.skipSpaces()
		if p.eof() {
			return out, p.errf("unexpected EOF while parsing nested map")
		}

		p.skipSpaces()
		//check if valid start
		if !p.consume('[') {
			return nil, p.errf("nested: expected '[' to start key")
		}
 
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
		return p.parseString() //is returned as a string
	case '{':
		return p.parseArray() //assume it's an array
	case '[':
		p.consume('[')
		v, err := p.parseValueSimple()
		if err != nil {
			return nil, err
		}
		if !p.consume(']') {
			return nil, p.errf("expected ']' after bracketed value")
		}
		return v, nil
	default:
		return p.parseIdentifierOrNumber()
	}
}

func (p *parser) parseArray() (interface{}, error) {
	if !p.consume('{') {
		return nil, p.errf("expected '{' to start array")
	}
	arr := make([]interface{}, 0)
	for {
		p.skipSpaces()
		if p.eof() {
			return nil, p.errf("unexpected EOF in array")
		}
		if p.peek() == '}' {
			p.consume('}')
			break
		}
		v, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		arr = append(arr, v)
		p.skipSpaces()
		if p.peek() == ',' {
			p.consume(',')
			p.skipSpaces()
			continue
		}
		if p.peek() == '}' {
			p.consume('}')
			break
		}
		return nil, p.errf("expected ',' or '}' in array")
	}
	return arr, nil
}

func (p *parser) parseString() (interface{}, error) {
	if !p.consume('"') && !p.consume('\'') {
		return nil, p.errf("expected '\"' or '\\'' to start string")
	}
	var sb strings.Builder
	escaped := false
	for !p.eof() {
		ch := p.next()
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
			escaped = false
			continue
		}
		if ch == '\\' {
			escaped = true
			continue
		}
		if ch == '"' {
			return sb.String(), nil
		}
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
		for !p.eof() && (unicode.IsDigit(rune(p.peek())) || p.peek() == '.') {
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
	for !p.eof() && (unicode.IsLetter(rune(p.peek())) || p.peek() == '_' || unicode.IsDigit(rune(p.peek()))) {
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
			return nil, p.errf("unexpected token at pos %d", p.pos)
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

	start := p.pos - 10
	if start < 0 {
		start = 0
	}

	end := p.pos + 10
	if end > p.n {
		end = p.n
	}

	snippet := p.s[start:end]

	errStr := fmt.Sprintf("%s (pos=%d) near %q", 
			msg, p.pos, snippet)

	return errors.New(errStr)
}
