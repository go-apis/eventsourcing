package es

import (
	"fmt"

	"github.com/contextcloud/eventstore/pkg/gcppubsub"
	"github.com/contextcloud/eventstore/pkg/natspubsub"
	"github.com/contextcloud/eventstore/pkg/pgdb"
)

type StreamConfig struct {
	Type   string
	PubSub *gcppubsub.Config
	Nats   *natspubsub.Config
}

type DataConfig struct {
	Type string
	Pg   *pgdb.Config
}

type ProviderConfig struct {
	ServiceName string
	Version     string

	Data   DataConfig
	Stream StreamConfig
}

type AggregateConfig struct {
	EntityOptions  []EntityOption
	CommandConfigs []*CommandConfig
	Handles        CommandHandles
}

func NewAggregateConfig(aggregate Aggregate, items ...interface{}) *AggregateConfig {
	entityOptions := NewEntityOptions(aggregate)
	var commandConfigs []*CommandConfig

	for _, item := range items {
		switch raw := item.(type) {
		case EntityOption:
			entityOptions = append(entityOptions, raw)
		case Command:
			cmdCfg := NewCommandConfig(raw)
			commandConfigs = append(commandConfigs, cmdCfg)
		default:
			panic(fmt.Sprintf("invalid item type: %T", item))
		}
	}

	return &AggregateConfig{
		EntityOptions:  entityOptions,
		CommandConfigs: commandConfigs,
	}
}

type EventHandlerConfig struct {
	handler      EventHandler
	eventConfigs []*EventConfig
}

type CommandHandlerConfig struct {
	handler        CommandHandler
	commandConfigs []*CommandConfig
}

type Config interface {
	GetProviderConfig() *ProviderConfig
	GetEntityConfigs() []*EntityConfig
	GetCommandConfigs() []*CommandConfig
	GetEventConfigs() []*EventConfig
	GetCommandHandlers() []*CommandHandlerConfig
	GetCommandHandlerMiddlewares() []CommandHandlerMiddleware
	GetEventHandlers() []*EventHandlerConfig
}

type config struct {
	providerConfig            *ProviderConfig
	entities                  []*EntityConfig
	commandConfigs            []*CommandConfig
	eventConfigs              []*EventConfig
	commandHandlers           []*CommandHandlerConfig
	commandHandlerMiddlewares []CommandHandlerMiddleware
	eventHandlers             []*EventHandlerConfig
}

func (c config) GetProviderConfig() *ProviderConfig {
	return c.providerConfig
}
func (c config) GetEntityConfigs() []*EntityConfig {
	return c.entities
}
func (c config) GetCommandConfigs() []*CommandConfig {
	return c.commandConfigs
}
func (c config) GetEventConfigs() []*EventConfig {
	return c.eventConfigs
}
func (c config) GetCommandHandlers() []*CommandHandlerConfig {
	return c.commandHandlers
}
func (c config) GetCommandHandlerMiddlewares() []CommandHandlerMiddleware {
	return c.commandHandlerMiddlewares
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
	c.commandConfigs = append(c.commandConfigs, cfg.CommandConfigs...)

	h := NewSourcedAggregateHandler(entityConfig.Name, cfg.Handles)
	c.commandHandlers = append(c.commandHandlers, &CommandHandlerConfig{
		handler:        h,
		commandConfigs: cfg.CommandConfigs,
	})
	return nil
}

// don't modify the object
func (c config) dynamic(agg Aggregate) error {
	handles := NewCommandHandles(agg)
	var commandConfigs []*CommandConfig
	for t := range handles {
		cmdConfig := NewCommandConfig(t)
		commandConfigs = append(commandConfigs, cmdConfig)
	}

	opts := NewEntityOptions(agg)
	aggregateConfig := &AggregateConfig{
		EntityOptions:  opts,
		CommandConfigs: commandConfigs,
		Handles:        handles,
	}
	return c.aggregate(aggregateConfig)
}

func (c *config) saga(s IsSaga) error {
	handles := NewSagaHandles(s)

	eventConfigs := []*EventConfig{}
	for t := range handles {
		evtConfig := NewEventConfig(t)
		eventConfigs = append(eventConfigs, evtConfig)
	}

	c.eventConfigs = append(c.eventConfigs, eventConfigs...)

	h := NewSagaEventHandler(handles, s)
	c.eventHandlers = append(c.eventHandlers, &EventHandlerConfig{
		handler:      h,
		eventConfigs: eventConfigs,
	})
	return nil
}

func (c *config) middleware(m CommandHandlerMiddleware) error {
	c.commandHandlerMiddlewares = append(c.commandHandlerMiddlewares, m)
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
	case CommandHandlerMiddleware:
		return c.middleware(raw)
	default:
		return fmt.Errorf("invalid item type: %T", item)
	}
}

func NewConfig(pcfg *ProviderConfig, items ...interface{}) (Config, error) {
	cfg := &config{
		providerConfig: pcfg,
	}

	for _, item := range items {
		if err := cfg.config(item); err != nil {
			return nil, err
		}
	}

	return cfg, nil
}
