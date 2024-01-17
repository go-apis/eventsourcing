package utils

import (
	"reflect"
	"strings"
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
	split := func(r rune) bool {
		return r == ';' || r == ','
	}
	return strings.FieldsFunc(tag, split)
}
