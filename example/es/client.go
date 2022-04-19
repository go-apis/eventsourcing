package es

import (
	"context"
	"eventstore/example/es/utils"
	"fmt"
	"reflect"
)

var ErrHandlerNotFound = fmt.Errorf("handler not found")

type Client interface {
	Dispatch(ctx context.Context, cmd Command) error
}

type client struct {
	handlers map[reflect.Type]CommandHandler
}

func (c *client) Dispatch(ctx context.Context, cmd Command) error {
	t := utils.GetElemType(cmd)
	h, ok := c.handlers[t]
	if !ok {
		return ErrHandlerNotFound
	}
	return h.Handle(ctx, cmd)
}

func NewClient() Client {
	return &client{}
}
