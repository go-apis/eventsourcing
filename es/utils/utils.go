package utils

import "reflect"

func GetElemType(source interface{}) reflect.Type {
	rawType := reflect.TypeOf(source)
	// source is a pointer, convert to its value
	if rawType.Kind() == reflect.Ptr {
		rawType = rawType.Elem()
	}
	return rawType
}
