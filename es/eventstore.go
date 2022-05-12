package es

import "reflect"

type EventStore interface {
	Dispatcher
	Store

	Config(opt ...Option) error
}

type eventStore struct {
	*dispatcher
	Store
}

func (es *eventStore) Config(opt ...Option) error {
	opts := defaultOptions
	for _, o := range opt {
		o.apply(&opts)
	}

	for _, h := range opts.commandHandlers {
		if err := es.dispatcher.AddCommandHandler(h); err != nil {
			return err
		}
	}

	for _, h := range opts.eventHandlers {
		if err := es.dispatcher.AddEventHandler(h); err != nil {
			return err
		}
	}

	return nil
}

func NewEventStore(serviceName string, dsn string) (EventStore, error) {
	d := &dispatcher{
		commandHandlers: make(map[reflect.Type]CommandHandler),
	}

	store, err := NewStore(dsn, serviceName)
	if err != nil {
		return nil, err
	}

	return &eventStore{
		dispatcher: d,
		Store:      store,
	}, nil
}
