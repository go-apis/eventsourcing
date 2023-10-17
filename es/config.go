package es

import (
	"fmt"

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
