package es

import (
	"context"
	"fmt"
	"reflect"
)

type CommandHandlers map[reflect.Type]CommandHandler

func (hs CommandHandlers) Add(hs2 CommandHandlers) error {
	for k, v := range hs2 {
		if _, ok := hs[k]; ok {
			return fmt.Errorf("handler already exists: %s", k)
		}
		hs[k] = v
	}
	return nil
}

type CommandHandler interface {
	Handle(ctx context.Context, cmd Command) error
}
