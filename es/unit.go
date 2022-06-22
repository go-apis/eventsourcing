package es

import (
	"context"
	"eventstore/es/filters"
	"fmt"
	"reflect"

	"github.com/neilotoole/errgroup"
)

var ErrHandlerNotFound = fmt.Errorf("handler not found")
var ErrNotCommandHandler = fmt.Errorf("not a command handler")

type Unit interface {
	DispatchAsync(ctx context.Context, cmds ...Command) error
	Dispatch(ctx context.Context, cmds ...Command) error
	Commit(ctx context.Context) error
	Load(ctx context.Context, id string, aggregateName string, out interface{}) error
	Find(ctx context.Context, aggregateName string, filter filters.Filter, out interface{}) error
}

type unit struct {
	tx              Tx
	serviceName     string
	commandHandlers map[reflect.Type]CommandHandler
}

func (u *unit) DispatchAsync(ctx context.Context, cmds ...Command) error {
	numG, qSize := 8, 4
	ctx = SetTx(ctx, u.tx)
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
	ctx = SetTx(ctx, u.tx)
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

func (u *unit) Commit(ctx context.Context) error {
	return u.tx.Commit(ctx)
}

func (u *unit) Load(ctx context.Context, id string, aggregateName string, out interface{}) error {
	namespace := NamespaceFromContext(ctx)
	return u.tx.Load(ctx, u.serviceName, aggregateName, namespace, id, out)
}
func (u *unit) Find(ctx context.Context, aggregateName string, filter filters.Filter, out interface{}) error {
	namespace := NamespaceFromContext(ctx)
	return u.tx.Find(ctx, u.serviceName, aggregateName, namespace, filter, out)
}

func NewUnit(cfg Config, tx Tx) (Unit, error) {
	return &unit{
		tx:              tx,
		serviceName:     cfg.GetServiceName(),
		commandHandlers: cfg.GetCommandHandlers(),
	}, nil
}
