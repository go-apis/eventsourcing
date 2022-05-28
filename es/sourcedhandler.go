package es

import (
	"context"
	"reflect"
)

type aggregateHandler struct {
	sourcedStore SourcedStore
	name         string
	factory      AggregateFactory
	handles      map[reflect.Type]*CommandHandle
}

func (b *aggregateHandler) Handle(ctx context.Context, cmd Command) error {
	t := reflect.TypeOf(cmd)
	h, ok := b.handles[t]
	if !ok {
		return ErrNotCommandHandler
	}

	aggregateId := cmd.GetAggregateId()
	agg, err := b.factory()
	if err != nil {
		return err
	}

	// get the transaction
	tx, err := b.sourcedStore.GetTx(ctx)
	if err != nil {
		return err
	}

	if err := tx.Load(ctx, aggregateId, b.name, agg); err != nil {
		return err
	}

	if err := h.Handle(agg, ctx, cmd); err != nil {
		return err
	}

	evts, err := tx.Save(ctx, aggregateId, b.name, agg)
	if err != nil {
		return err
	}

	// if err := b.eventStore.dispatcher.PublishAsync(ctx, evts...); err != nil {
	// 	return err
	// }

	return nil
}

func NewSourcedAggregateHandler(sourcedStore SourcedStore, name string, factory AggregateFactory, handles CommandHandles) CommandHandler {
	return &aggregateHandler{
		sourcedStore: sourcedStore,
		name:         name,
		factory:      factory,
		handles:      handles,
	}
}
