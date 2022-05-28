package es

import (
	"context"
	"reflect"
)

type aggregateHandler struct {
	sourcedStore SourcedStore
	handles      map[reflect.Type]*CommandHandle
}

func (b *aggregateHandler) Handle(ctx context.Context, cmd Command) error {
	t := reflect.TypeOf(cmd)
	h, ok := b.handles[t]
	if !ok {
		return ErrNotCommandHandler
	}

	aggregateId := cmd.GetAggregateId()
	agg, err := b.sourcedStore.Load(ctx, aggregateId)
	if err != nil {
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

func NewSourcedAggregateHandler(sourcedStore SourcedStore, handles CommandHandles) CommandHandler {
	return &aggregateHandler{
		sourcedStore: sourcedStore,
		handles:      handles,
	}
}
