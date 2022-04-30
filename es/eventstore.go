package es

import (
	"fmt"
)

var ErrNotCommandHandler = fmt.Errorf("not a command handler")

type EventStore interface {
	Dispatcher
	Store
}

type eventStore struct {
	Dispatcher
	Store
}

func NewEventStore(opt ...Option) (EventStore, error) {
	opts := defaultOptions
	for _, o := range opt {
		o.apply(&opts)
	}

	store, err := NewStore(opts.url, opts.serviceName)
	if err != nil {
		return nil, err
	}

	// TODO: how do we initialize things? ie create db entities?

	// create the handlers.
	handlers := make(CommandHandlers)
	for _, h := range opts.handlers {
		hs, err := BuildCommandHandlers(store, h)
		if err != nil {
			return nil, err
		}

		if err := handlers.Add(hs); err != nil {
			return nil, err
		}
	}

	dispatcher, err := NewDispatcher(handlers)
	if err != nil {
		return nil, err
	}

	return &eventStore{
		Dispatcher: dispatcher,
		Store:      store,
	}, nil
}

func BuildCommandHandlers(store Store, h interface{}) (CommandHandlers, error) {
	switch impl := h.(type) {
	case Aggregate:
		return NewBaseAggregateHandlers(store, impl)
	default:
		return nil, ErrNotCommandHandler
	}
}
