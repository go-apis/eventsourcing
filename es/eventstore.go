package es

import (
	"reflect"

	"github.com/jinzhu/copier"
)

type EventStore interface {
	Dispatcher
	Data

	AddCommandHandler(handlers ...interface{}) error
	AddEventHandler(handlers ...interface{}) error
}

type eventStore struct {
	*dispatcher
	Data

	serviceName string
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
			s := NewSourcedStore(e, name, factory)
			h := NewSourcedAggregateHandler(s, handles)
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

func NewEventStore(data Data, serviceName string) EventStore {
	d := &dispatcher{
		commandHandlers: make(map[reflect.Type]CommandHandler),
		eventHandlers:   make(map[reflect.Type][]EventHandler),
	}

	return &eventStore{
		dispatcher:  d,
		Data:        data,
		serviceName: serviceName,
	}
}
