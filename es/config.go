package es

import (
	"fmt"
	"reflect"

	"github.com/jinzhu/copier"
)

type EntityConfig interface {
	IsEntityConfig()
	GetOptions() EntityOptions
	GetType() reflect.Type
}
type entityConfig struct {
	options    EntityOptions
	entityType reflect.Type
}

func (cfg entityConfig) IsEntityConfig() {}
func (cfg entityConfig) GetOptions() EntityOptions {
	return cfg.options
}
func (cfg entityConfig) GetType() reflect.Type {
	return cfg.entityType
}
func NewEntityConfig() EntityConfig {
	return &entityConfig{}
}

type EventHandlerConfig interface {
	IsEventHandlerConfig()
	GetHandler() EventHandler
	GetEvents() []reflect.Type
}
type eventHandlerConfig struct {
	handler EventHandler
	events  []reflect.Type
}

func (cfg eventHandlerConfig) IsEventHandlerConfig() {}
func (cfg eventHandlerConfig) GetHandler() EventHandler {
	return cfg.handler
}
func (cfg eventHandlerConfig) GetEvents() []reflect.Type {
	return cfg.events
}
func NewEventHandlerConfig() EventHandlerConfig {
	return &eventHandlerConfig{}
}

type CommandHandlerConfig interface {
	IsCommandHandlerConfig()
	GetHandler() CommandHandler
	GetCommands() []reflect.Type
}
type commandHandlerConfig struct {
	handler  CommandHandler
	commands []reflect.Type
}

func (cfg commandHandlerConfig) IsCommandHandlerConfig() {}
func (cfg commandHandlerConfig) GetHandler() CommandHandler {
	return cfg.handler
}
func (cfg commandHandlerConfig) GetCommands() []reflect.Type {
	return cfg.commands
}
func NewCommandHandlerConfig() CommandHandlerConfig {
	return &commandHandlerConfig{}
}

type Config interface {
	GetServiceName() string
	GetEntities() []EntityConfig
	GetCommandHandlers() []CommandHandlerConfig
	GetEventHandlers() []EventHandlerConfig
}

type config struct {
	serviceName string

	entities        []EntityConfig
	commandHandlers []CommandHandlerConfig
	eventHandlers   []EventHandlerConfig
}

func (c config) GetServiceName() string {
	return c.serviceName
}

func (c config) GetEntities() []EntityConfig {
	return c.entities
}
func (c config) GetCommandHandlers() []CommandHandlerConfig {
	return c.commandHandlers
}

func (c config) GetEventHandlers() []EventHandlerConfig {
	return c.eventHandlers
}

func (c *config) sourced(agg AggregateSourced) error {
	t := reflect.TypeOf(agg)
	handles := NewCommandHandles(t)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// TODO read tags from field here!

	name := t.String()
	factory := func() (Entity, error) {
		out := reflect.New(t).Interface().(Entity)
		if err := copier.Copy(out, agg); err != nil {
			return nil, err
		}
		return out, nil
	}
	opts := []EntityOption{
		EntityName(name),
		EntityFactory(factory),
	}
	c.entities = append(c.entities, &entityConfig{
		options:    NewEntityOptions(opts),
		entityType: t,
	})

	commands := []reflect.Type{}
	for t := range handles {
		commands = append(commands, t)
	}
	h := NewSourcedAggregateHandler(name, handles)

	c.commandHandlers = append(c.commandHandlers, &commandHandlerConfig{
		handler:  h,
		commands: commands,
	})
	return nil
}

func (c *config) saga(s IsSaga) error {
	t := reflect.TypeOf(s)
	handles := NewSagaHandles(t)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	events := []reflect.Type{}
	for t := range handles {
		events = append(events, t)
	}

	h := NewSagaEventHandler(handles, s)
	c.eventHandlers = append(c.eventHandlers, &eventHandlerConfig{
		handler: h,
		events:  events,
	})
	return nil
}

func (c *config) commandHandlerConfig(a CommandHandlerConfig) error {
	c.commandHandlers = append(c.commandHandlers, a)
	return nil
}

func (c *config) eventHandlerConfig(a EventHandlerConfig) error {
	c.eventHandlers = append(c.eventHandlers, a)
	return nil
}

func (c *config) config(item interface{}) error {
	switch raw := item.(type) {
	case CommandHandlerConfig:
		return c.commandHandlerConfig(raw)
	case EventHandlerConfig:
		return c.eventHandlerConfig(raw)
	case IsSaga:
		return c.saga(raw)
	case AggregateSourced:
		return c.sourced(raw)
	default:
		return fmt.Errorf("invalid item type: %T", item)
	}
}

func NewConfig(serviceName string, items ...interface{}) (Config, error) {
	cfg := &config{
		serviceName: serviceName,
	}

	for _, item := range items {
		if err := cfg.config(item); err != nil {
			return nil, err
		}
	}

	return cfg, nil
}
