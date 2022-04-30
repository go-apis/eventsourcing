package es

import (
	"context"
	"fmt"
	"reflect"

	"github.com/neilotoole/errgroup"
)

var ErrHandlerNotFound = fmt.Errorf("handler not found")

type Dispatcher interface {
	Dispatch(ctx context.Context, cmds ...Command) error
}

type dispatcher struct {
	handlers CommandHandlers
}

func (c *dispatcher) Dispatch(ctx context.Context, cmds ...Command) error {
	numG, qSize := 8, 4
	g, ctx := errgroup.WithContextN(ctx, numG, qSize)

	for _, cmd := range cmds {
		in := cmd

		g.Go(func() error {
			t := reflect.TypeOf(in)
			h, ok := c.handlers[t]
			if !ok {
				return ErrHandlerNotFound
			}
			return h.Handle(ctx, in)
		})
	}

	return g.Wait()
}

func NewDispatcher(handlers CommandHandlers) (Dispatcher, error) {
	return &dispatcher{
		handlers: handlers,
	}, nil
}
