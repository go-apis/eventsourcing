package es

import (
	"context"
	"reflect"
)

type aggregateHandler struct {
	sourcedStore SourcedStore
	handles      map[reflect.Type]*CommandHandle
	factory      SourcedAggregateFactory
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

	if err := b.sourcedStore.Load(ctx, aggregateId, agg); err != nil {
		return err
	}

	if err := h.Handle(agg, ctx, cmd); err != nil {
		return err
	}

	if err := b.sourcedStore.Save(ctx, aggregateId, agg); err != nil {
		return err
	}

	return nil
}

func NewSourcedAggregateHandler(sourcedStore SourcedStore, handles CommandHandles, factory SourcedAggregateFactory) CommandHandler {
	return &aggregateHandler{
		sourcedStore: sourcedStore,
		handles:      handles,
		factory:      factory,
	}
}
