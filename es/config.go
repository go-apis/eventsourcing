package es

import (
	"fmt"
	"reflect"

	"github.com/contextcloud/eventstore/pkg/gcppubsub"
	"github.com/contextcloud/eventstore/pkg/natspubsub"
	"github.com/contextcloud/goutils/xgorm"
)

type StreamConfig struct {
	Type   string
	PubSub *gcppubsub.Config
	Nats   *natspubsub.Config
	AWS    *AwsSnsConfig
}

type AwsSnsConfig struct {
	Region              string
	TopicARN            string
	QueueName           string
	MaxNumberOfMessages int
	WaitTimeSeconds     int
}

type DataConfig struct {
	Type  string
	Pg    *xgorm.DbConfig
	Reset bool
}

type ProviderConfig struct {
	Service string
	Version string

	Data   DataConfig
	Stream StreamConfig
}

type AggregateConfig struct {
	EntityOptions  []EntityOption
	CommandConfigs []*CommandConfig
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
	GetEntityConfigs() map[string]*EntityConfig
	GetCommandConfigs() map[string]*CommandConfig
	GetEventConfigs() map[string]*EventConfig

	GetReplayHandler(entityName string) CommandHandler
	GetCommandHandlers() map[reflect.Type]CommandHandler
	GetEventHandlers() map[reflect.Type][]EventHandler
}

func NewConfig(pcfg *ProviderConfig, items ...interface{}) (Config, error) {
	b := NewBuilder().
		SetProviderConfig(pcfg)

	for _, item := range items {
		switch raw := item.(type) {
		case IsCommandHandler:
			b.AddCommandHandler(raw)
			continue
		case IsSaga:
			b.AddSaga(raw)
			continue
		case IsProjector:
			b.AddProjector(raw)
			continue
		case Aggregate:
			b.AddAggregate(raw)
			continue
		case Entity:
			b.AddEntity(raw)
			continue
		case *AggregateConfig:
			b.AddAggregateConfig(raw)
			continue
		case CommandHandlerMiddleware:
			b.AddMiddleware(raw)
			continue
		case EventPublish:
			b.AddEventPublish(raw)
		case EventPublished:
			b.AddEventPublished(raw)
		default:
			return nil, fmt.Errorf("invalid item type: %T", item)
		}
	}

	return b.Build()
}
