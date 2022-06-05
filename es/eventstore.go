package es

import (
	"context"
	"reflect"

	"github.com/jinzhu/copier"
)

type EventStore interface {
	AddCommandHandler(handlers ...interface{}) error
	AddEventHandler(handlers ...interface{}) error

	NewTx(ctx context.Context) (Transaction, error)
}

type eventStore struct {
	*dispatcher

	client Client
}

func (e *eventStore) NewTx(ctx context.Context) (Transaction, error) {
	tx, err := e.client.NewTx(ctx)
	if err != nil {
		return nil, err
	}

	return NewTransaction(tx)
}

func (e *eventStore) AddCommandHandler(handlers ...interface{}) error {
	for _, h := range handlers {
		t := reflect.TypeOf(h)
		handles := NewCommandHandles(t)
		for t.Kind() == reflect.Ptr {
			t = t.Elem()
		}

		switch impl := h.(type) {
		case SourcedAggregate:
			name := t.String()
			factory := func() (SourcedAggregate, error) {
				agg := reflect.New(t).Interface().(SourcedAggregate)
				if err := copier.Copy(agg, impl); err != nil {
					return nil, err
				}
				return agg, nil
			}
			s := e.client.NewSourcedStore(e.dispatcher, name)
			h := NewSourcedAggregateHandler(s, handles, factory)
			for _, ch := range handles {
				e.commandHandlers[ch.commandType] = h
			}
		default:
			return ErrNotCommandHandler
		}
	}
	return nil
}

func (e *eventStore) AddEventHandler(handlers ...interface{}) error {
	for _, h := range handlers {
		t := reflect.TypeOf(h)
		handles := NewEventHandles(t)
		for t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		name := t.String()
		eh := NewEventHandler(name, h, handles)
		for _, ch := range handles {
			e.eventHandlers[ch.eventType] = append(e.eventHandlers[ch.eventType], eh)
		}
	}
	return nil
}

func NewEventStore(client Client) EventStore {
	d := &dispatcher{
		commandHandlers: make(map[reflect.Type]CommandHandler),
		eventHandlers:   make(map[reflect.Type][]EventHandler),
	}

	return &eventStore{
		dispatcher: d,
		client:     client,
	}
}
