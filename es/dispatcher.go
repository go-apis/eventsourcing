package es

import (
	"context"
	"fmt"
	"reflect"
)

var ErrHandlerNotFound = fmt.Errorf("handler not found")

type Dispatcher interface {
	Dispatch(ctx context.Context, cmd Command) error
}

type dispatcher struct {
	handlers CommandHandlers
}

func (c *dispatcher) Dispatch(ctx context.Context, cmd Command) error {
	t := reflect.TypeOf(cmd)
	h, ok := c.handlers[t]
	if !ok {
		return ErrHandlerNotFound
	}
	return h.Handle(ctx, cmd)
}

func NewDispatcher(handlers CommandHandlers) (Dispatcher, error) {
	return &dispatcher{
		handlers: handlers,
	}, nil
}
