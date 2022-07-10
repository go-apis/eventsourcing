package es

import (
	"context"
	"fmt"
	"reflect"

	"github.com/contextcloud/eventstore/es/filters"

	"github.com/neilotoole/errgroup"
)

var ErrHandlerNotFound = fmt.Errorf("handler not found")
var ErrNotCommandHandler = fmt.Errorf("not a command handler")

type Unit interface {
	DispatchAsync(ctx context.Context, cmds ...Command) error
	Dispatch(ctx context.Context, cmds ...Command) error

	GetData() Data
	NewTx(ctx context.Context) (Tx, error)
	Load(ctx context.Context, id string, aggregateName string, out interface{}) error
	Find(ctx context.Context, aggregateName string, filter filters.Filter, out interface{}) error
	Count(ctx context.Context, aggregateName string, filter filters.Filter) (int, error)
}

type unit struct {
	data            Data
	serviceName     string
	commandHandlers map[reflect.Type]CommandHandler
}

func (u *unit) DispatchAsync(ctx context.Context, cmds ...Command) error {
	numG, qSize := 8, 4
	ctx = SetUnit(ctx, u)
	g, ctx := errgroup.WithContextN(ctx, numG, qSize)

	for _, cmd := range cmds {
		in := cmd

		g.Go(func() error {
			t := reflect.TypeOf(in)
			h, ok := u.commandHandlers[t]
			if !ok {
				return ErrHandlerNotFound
			}
			return h.Handle(ctx, in)
		})
	}

	return g.Wait()
}

func (u *unit) Dispatch(ctx context.Context, cmds ...Command) error {
	ctx = SetUnit(ctx, u)

	for _, cmd := range cmds {
		t := reflect.TypeOf(cmd)
		h, ok := u.commandHandlers[t]
		if !ok {
			return ErrHandlerNotFound
		}
		if err := h.Handle(ctx, cmd); err != nil {
			return err
		}
	}
	return nil
}

func (u *unit) GetData() Data {
	return u.data
}

func (u *unit) NewTx(ctx context.Context) (Tx, error) {
	return u.data.Begin(ctx)
}

func (u *unit) Load(ctx context.Context, id string, aggregateName string, out interface{}) error {
	namespace := NamespaceFromContext(ctx)
	return u.data.Load(ctx, u.serviceName, aggregateName, namespace, id, out)
}
func (u *unit) Find(ctx context.Context, aggregateName string, filter filters.Filter, out interface{}) error {
	namespace := NamespaceFromContext(ctx)
	return u.data.Find(ctx, u.serviceName, aggregateName, namespace, filter, out)
}

func (u *unit) Count(ctx context.Context, aggregateName string, filter filters.Filter) (int, error) {
	namespace := NamespaceFromContext(ctx)
	return u.data.Count(ctx, u.serviceName, aggregateName, namespace, filter)
}

func newUnit(cfg Config, data Data) (Unit, error) {
	return &unit{
		data:            data,
		serviceName:     cfg.GetServiceName(),
		commandHandlers: cfg.GetCommandHandlers(),
	}, nil
}
