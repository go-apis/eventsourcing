package es

import (
	"fmt"
	"reflect"
)

type AggregateConfig struct {
	EntityOptions []EntityOption
	Commands      []reflect.Type
	Handles       CommandHandles
}

func NewAggregateConfig(aggregate Aggregate, items ...interface{}) *AggregateConfig {
	entityOptions := NewEntityOptions(aggregate)
	var commands []reflect.Type

	for _, item := range items {
		switch raw := item.(type) {
		case EntityOption:
			entityOptions = append(entityOptions, raw)
		case Command:
			t := reflect.TypeOf(raw)
			commands = append(commands, t)
		default:
			panic(fmt.Sprintf("invalid item type: %T", item))
		}
	}

	return &AggregateConfig{
		EntityOptions: entityOptions,
		Commands:      commands,
	}
}

// EventOptions represents the configuration options
// for the event.
type EventConfig struct {
	Name    string
	Type    reflect.Type
	Factory func() (interface{}, error)
}

type EventHandlerConfig struct {
	handler EventHandler
	events  []reflect.Type
}

type CommandHandlerConfig struct {
	handler  CommandHandler
	commands []reflect.Type
}

type Config interface {
	GetServiceName() string
	GetServiceVersion() string
	GetEntities() []*EntityConfig
	GetCommandHandlers() []*CommandHandlerConfig
	GetEventHandlers() []*EventHandlerConfig
}

type config struct {
	serviceName    string
	serviceVersion string

	entities        []*EntityConfig
	commandHandlers []*CommandHandlerConfig
	eventHandlers   []*EventHandlerConfig
}

func (c config) GetServiceName() string {
	return c.serviceName
}
func (c config) GetServiceVersion() string {
	return c.serviceVersion
}
func (c config) GetEntities() []*EntityConfig {
	return c.entities
}
func (c config) GetCommandHandlers() []*CommandHandlerConfig {
	return c.commandHandlers
}
func (c config) GetEventHandlers() []*EventHandlerConfig {
	return c.eventHandlers
}

func (c *config) aggregate(cfg *AggregateConfig) error {
	entityConfig, err := NewEntityConfig(cfg.EntityOptions)
	if err != nil {
		return err
	}

	c.entities = append(c.entities, entityConfig)

	h := NewSourcedAggregateHandler(entityConfig.Name, cfg.Handles)
	c.commandHandlers = append(c.commandHandlers, &CommandHandlerConfig{
		handler:  h,
		commands: cfg.Commands,
	})
	return nil
}

func (c *config) dynamic(agg Aggregate) error {
	handles := NewCommandHandles(agg)
	commandTypes := []reflect.Type{}
	for commandType := range handles {
		commandTypes = append(commandTypes, commandType)
	}

	opts := NewEntityOptions(agg)
	aggregateConfig := &AggregateConfig{
		EntityOptions: opts,
		Commands:      commandTypes,
		Handles:       handles,
	}
	return c.aggregate(aggregateConfig)
}

func (c *config) saga(s IsSaga) error {
	handles := NewSagaHandles(s)

	events := []reflect.Type{}
	for t := range handles {
		events = append(events, t)
	}

	h := NewSagaEventHandler(handles, s)
	c.eventHandlers = append(c.eventHandlers, &EventHandlerConfig{
		handler: h,
		events:  events,
	})
	return nil
}

func (c *config) config(item interface{}) error {
	switch raw := item.(type) {
	case IsSaga:
		return c.saga(raw)
	case Aggregate:
		return c.dynamic(raw)
	case *AggregateConfig:
		return c.aggregate(raw)
	default:
		return fmt.Errorf("invalid item type: %T", item)
	}
}

func NewConfig(serviceName string, serviceVersion string, items ...interface{}) (Config, error) {
	cfg := &config{
		serviceName:    serviceName,
		serviceVersion: serviceVersion,
	}

	for _, item := range items {
		if err := cfg.config(item); err != nil {
			return nil, err
		}
	}

	return cfg, nil
}
