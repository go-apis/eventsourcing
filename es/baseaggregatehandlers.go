package es

import (
	"context"
	"reflect"
)

type wrapped struct {
	Aggregate interface{}
	Version   int
}

type baseAggregateHandler struct {
	store    Store
	typeName string
	factory  func() interface{}
	handles  map[reflect.Type]*CommandHandle
}

func (b *baseAggregateHandler) Handle(ctx context.Context, cmd Command) error {
	t := reflect.TypeOf(cmd)
	h, ok := b.handles[t]
	if !ok {
		return ErrNotCommandHandler
	}

	aggregateId := cmd.GetAggregateId()
	// TODO: get the namespace
	namespace := "default"

	agg := b.factory()
	if err := b.store.LoadSnapshot(ctx, namespace, aggregateId, agg); err != nil {
		return err
	}

	// get the events
	events, err := b.store.LoadEvents(ctx, namespace, aggregateId, 0)
	if err != nil {
		return err
	}

	for _, evt := range events {

	}

	// todo load it!.
	return h.Handle(agg, ctx, cmd)
}

func NewBaseAggregateHandlers(store Store, agg Aggregate) (CommandHandlers, error) {
	t := reflect.TypeOf(agg)
	handles := NewCommandHandles(t)

	raw := t
	if raw.Kind() == reflect.Ptr {
		raw = raw.Elem()
	}
	factory := func() interface{} {
		return reflect.New(raw).Interface()
	}
	handler := &baseAggregateHandler{
		store:   store,
		factory: factory,
		handles: handles,
	}

	handlers := make(CommandHandlers)
	for h := range handles {
		handlers[h] = handler
	}
	return handlers, nil
}
