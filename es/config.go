package es

import (
	"fmt"
	"reflect"

	"github.com/jinzhu/copier"
)

type Config interface {
	GetServiceName() string
	GetEntities() map[reflect.Type]*Entity
	GetCommandHandlers() map[reflect.Type]CommandHandler
}

type config struct {
	serviceName string

	entities        map[reflect.Type]*Entity
	commandHandlers map[reflect.Type]CommandHandler
	eventHandlers   map[reflect.Type][]EventHandler
}

func (c *config) GetServiceName() string {
	return c.serviceName
}
func (c *config) GetEntities() map[reflect.Type]*Entity {
	return c.entities
}
func (c *config) GetCommandHandlers() map[reflect.Type]CommandHandler {
	return c.commandHandlers
}

func (c *config) sourced(agg SourcedAggregate) error {
	t := reflect.TypeOf(agg)
	handles := NewCommandHandles(t)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	name := t.String()
	factory := func() (SourcedAggregate, error) {
		agg := reflect.New(t).Interface().(SourcedAggregate)
		if err := copier.Copy(agg, agg); err != nil {
			return nil, err
		}
		return agg, nil
	}

	pub := NewPublisher(c.eventHandlers)
	store := NewSourcedStore(pub, c.serviceName, name)

	for t, handle := range handles {
		h := NewSourcedAggregateHandler(handle, factory, store)
		c.commandHandlers[t] = h
	}

	c.entities[t] = &Entity{
		ServiceName:   c.serviceName,
		AggregateType: name,
		Data:          agg,
	}
	return nil
}

func (c *config) saga(s Saga) error {
	t := reflect.TypeOf(s)
	handles := NewEventHandles(t)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	// name := t.String()
	factory := func() (Saga, error) {
		agg := reflect.New(t).Interface().(Saga)
		if err := copier.Copy(agg, agg); err != nil {
			return nil, err
		}
		return agg, nil
	}

	for t, handle := range handles {
		h := NewSagaEventHandler(handle, factory)
		c.eventHandlers[t] = append(c.eventHandlers[t], h)
	}

	return nil
}

func (c *config) config(item interface{}) error {
	switch raw := item.(type) {
	case Saga:
		return c.saga(raw)
	case SourcedAggregate:
		return c.sourced(raw)
	default:
		return fmt.Errorf("Invalid item type: %T", item)
	}
}

func NewConfig(serviceName string, items ...interface{}) (Config, error) {
	cfg := &config{
		serviceName:     serviceName,
		entities:        make(map[reflect.Type]*Entity),
		commandHandlers: map[reflect.Type]CommandHandler{},
		eventHandlers:   map[reflect.Type][]EventHandler{},
	}

	for _, item := range items {
		if err := cfg.config(item); err != nil {
			return nil, err
		}
	}

	return cfg, nil
}
