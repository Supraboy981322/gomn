package gomn

import (
	"os"
)

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
