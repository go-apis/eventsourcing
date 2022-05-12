package es

import (
	"context"
	"fmt"
	"reflect"

	"github.com/jinzhu/copier"
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
	eventHandlers   map[reflect.Type]EventHandler
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

func (r *dispatcher) AddCommandHandler(h interface{}) error {
	// TODO support non reflection based config

	t := reflect.TypeOf(h)
	handles := NewCommandHandles(t)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	switch impl := h.(type) {
	case SourcedAggregate:
		name := t.String()
		factory := func() (Aggregate, error) {
			agg := reflect.New(t).Interface().(Aggregate)
			if err := copier.Copy(agg, impl); err != nil {
				return nil, err
			}
			return agg, nil
		}
		h := NewSourcedAggregateHandler(name, factory, handles)
		for _, ch := range handles {
			r.commandHandlers[ch.commandType] = h
		}
		return nil
	default:
		return ErrNotCommandHandler
	}
}

func (r *dispatcher) AddEventHandler(h interface{}) error {
	// TODO support non reflection based config

	t := reflect.TypeOf(h)
	handles := NewEventHandles(t)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	name := t.String()
	eh := NewEventHandler(name, h, handles)
	for _, ch := range handles {
		r.eventHandlers[ch.eventType] = eh
	}

	return nil
}
