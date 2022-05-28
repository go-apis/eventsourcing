package es

import (
	"context"
	"fmt"
	"reflect"

	"github.com/neilotoole/errgroup"
)

var ErrHandlerNotFound = fmt.Errorf("handler not found")
var ErrNotCommandHandler = fmt.Errorf("not a command handler")
var ErrNotEventHandler = fmt.Errorf("not a event handler")

type Dispatcher interface {
	DispatchAsync(ctx context.Context, cmds ...Command) error
	Dispatch(ctx context.Context, cmds ...Command) error
}

type dispatcher struct {
	commandHandlers map[reflect.Type]CommandHandler
	eventHandlers   map[reflect.Type][]EventHandler
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

func (c *dispatcher) PublishAsync(ctx context.Context, evts ...Event) error {
	numG, qSize := 8, 4
	g, ctx := errgroup.WithContextN(ctx, numG, qSize)

	for _, evt := range evts {
		t := reflect.TypeOf(evt.Data)
		handlers := c.eventHandlers[t]

		for _, h := range handlers {
			d := evt
			in := h
			g.Go(func() error {
				return in.Handle(ctx, d, d.Data)
			})
		}
	}

	return g.Wait()
}
