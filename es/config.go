package es

import (
	"context"
	"fmt"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/go-apis/utils/xgorm"
)

type MemoryBusPubSub interface {
	Publish(topic string, messages ...*message.Message) error
	Subscribe(ctx context.Context, topic string) (<-chan *message.Message, error)
}

type StreamConfig struct {
	Type   string
	PubSub *GcpPubSubConfig
	Nats   *NatsConfig
	AWS    *AwsSnsConfig
	Memory *MemoryBusConfig
}

type GcpPubSubConfig struct {
	ProjectId string
	TopicId   string
}

type NatsConfig struct {
	Url     string
	Subject string
}

type AwsSnsConfig struct {
	Profile   string
	Region    string
	TopicArn  string
	QueueName string
}

type MemoryBusConfig struct {
	Topic  string
	PubSub MemoryBusPubSub
}

type SqliteConfig struct {
	File   string
	Memory bool
}

type DataConfig struct {
	Type   string
	Pg     *xgorm.DbConfig
	Sqlite *SqliteConfig
	Reset  bool
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
