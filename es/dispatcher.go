package es

import (
	"context"
	"fmt"
	"reflect"

	"github.com/neilotoole/errgroup"
)

var ErrHandlerNotFound = fmt.Errorf("handler not found")
var ErrNotCommandHandler = fmt.Errorf("not a command handler")

type Dispatcher interface {
	DispatchAsync(ctx context.Context, cmds ...Command) error
	Dispatch(ctx context.Context, cmds ...Command) error
}

type dispatcher struct {
	commandHandlers map[reflect.Type]CommandHandler
}

func (c *dispatcher) DispatchAsync(ctx context.Context, cmds ...Command) error {
	numG, qSize := 8, 4
	g, ctx := errgroup.WithContextN(ctx, numG, qSize)

	for _, cmd := range cmds {
		in := cmd

		g.Go(func() error {
			t := reflect.TypeOf(in)
			h, ok := c.commandHandlers[t]
			if !ok {
				return ErrHandlerNotFound
			}
			return h.Handle(ctx, in)
		})
	}

	return g.Wait()
}

func (c *dispatcher) Dispatch(ctx context.Context, cmds ...Command) error {
	for _, cmd := range cmds {
		t := reflect.TypeOf(cmd)
		h, ok := c.commandHandlers[t]
		if !ok {
			return ErrHandlerNotFound
		}
		if err := h.Handle(ctx, cmd); err != nil {
			return err
		}
	}
	return nil
}
