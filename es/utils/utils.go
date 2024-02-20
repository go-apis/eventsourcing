package utils

import (
	"reflect"
)

func GetElemType(source interface{}) reflect.Type {
	rawType := reflect.TypeOf(source)
	// source is a pointer, convert to its value
	if rawType.Kind() == reflect.Ptr {
		rawType = rawType.Elem()
	}
	return rawType
}

func GetTypeName(source interface{}) string {
	raw := GetElemType(source)
	return raw.Name()
}

func SplitTag(tag string) []string {
	var token []byte
	var tokens []string
	for i := 0; i < len(tag); i++ {
		if tag[i] == ',' || tag[i] == ';' {
			tokens = append(tokens, string(token))
			token = token[:0]
			continue
		} else if tag[i] == '\\' && i+1 < len(tag) {
			i++
			token = append(token, tag[i])
		} else {
			token = append(token, tag[i])
		}
	}
	tokens = append(tokens, string(token))
	return tokens
}
