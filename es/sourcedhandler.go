package es

import (
	"context"
	"reflect"
)

type aggregateHandler struct {
	name    string
	factory AggregateFactory
	handles map[reflect.Type]*CommandHandle
}

func (b *aggregateHandler) Handle(ctx context.Context, cmd Command) error {
	t := reflect.TypeOf(cmd)
	h, ok := b.handles[t]
	if !ok {
		return ErrNotCommandHandler
	}

	aggregateId := cmd.GetAggregateId()

	// get the transaction
	tx, err := TransactionCtx(ctx)
	if err != nil {
		return err
	}

	agg, err := b.factory()
	if err != nil {
		return err
	}

	if err := tx.Load(ctx, aggregateId, b.name, agg); err != nil {
		return err
	}

	// todo load it!.
	if err := h.Handle(agg, ctx, cmd); err != nil {
		return err
	}

	if _, err := tx.Save(ctx, aggregateId, b.name, agg); err != nil {
		return err
	}

	return nil
}

func NewSourcedAggregateHandler(name string, factory AggregateFactory, handles CommandHandles) CommandHandler {
	return &aggregateHandler{
		name:    name,
		factory: factory,
		handles: handles,
	}
}
