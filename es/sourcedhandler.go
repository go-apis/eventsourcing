package es

import (
	"context"
)

type aggregateHandler struct {
	handle       *CommandHandle
	factory      SourcedAggregateFactory
	sourcedStore SourcedStore
}

func (b *aggregateHandler) Handle(ctx context.Context, cmd Command) error {
	namespace := NamespaceFromContext(ctx)

	aggregateId := cmd.GetAggregateId()
	agg, err := b.factory()
	if err != nil {
		return err
	}

	switch a := agg.(type) {
	case SetAggregate:
		a.SetId(aggregateId, namespace)
	}

	if err := b.sourcedStore.Load(ctx, aggregateId, namespace, agg); err != nil {
		return err
	}

	if err := b.handle.Handle(agg, ctx, cmd); err != nil {
		return err
	}

	if err := b.sourcedStore.Save(ctx, aggregateId, namespace, agg); err != nil {
		return err
	}

	return nil
}

func NewSourcedAggregateHandler(handle *CommandHandle, factory SourcedAggregateFactory, sourcedStore SourcedStore) CommandHandler {
	return &aggregateHandler{
		handle:       handle,
		factory:      factory,
		sourcedStore: sourcedStore,
	}
}
