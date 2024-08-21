package es

import (
	"fmt"
	"reflect"
)

// CommandFactory for creating a command
type CommandFactory func() (Command, error)

// CommandConfig information for a command
type CommandConfig struct {
	Name    string
	Type    reflect.Type
	Factory CommandFactory
}

func NewCommandConfig(obj interface{}) *CommandConfig {
	var t reflect.Type

	switch raw := obj.(type) {
	case Command:
		t = reflect.TypeOf(raw)
	case reflect.Type:
		t = raw
	default:
		panic(fmt.Errorf("invalid type %v", raw))
	}

	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	name := t.Name()
	factory := func() (Command, error) {
		out := reflect.New(t).Interface().(Command)
		return out, nil
	}

	return &CommandConfig{
		Name:    name,
		Type:    t,
		Factory: factory,
	}
}
