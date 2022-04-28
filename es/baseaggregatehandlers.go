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
	factory  func() SourcedAggregate
	handles  map[reflect.Type]*CommandHandle
}

func (b *baseAggregateHandler) Handle(ctx context.Context, cmd Command) error {
	t := reflect.TypeOf(cmd)
	h, ok := b.handles[t]
	if !ok {
		return ErrNotCommandHandler
	}

	aggregateId := cmd.GetAggregateId()

	agg := b.factory()
	if err := b.store.Load(ctx, aggregateId, b.typeName, agg); err != nil {
		return err
	}

	// todo load it!.
	if err := h.Handle(agg, ctx, cmd); err != nil {
		return err
	}

	// save the events.
	agg
}

func NewBaseAggregateHandlers(store Store, agg Aggregate) (CommandHandlers, error) {
	t := reflect.TypeOf(agg)
	handles := NewCommandHandles(t)

	raw := t
	if raw.Kind() == reflect.Ptr {
		raw = raw.Elem()
	}
	factory := func() SourcedAggregate {
		return reflect.New(raw).Interface().(SourcedAggregate)
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
