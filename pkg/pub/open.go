package pub

import (
	"context"
	"fmt"

	"cloud.google.com/go/pubsub"
)

type Pub struct {
	cli   *pubsub.Client
	topic *pubsub.Topic
}

func (p *Pub) Publish(ctx context.Context, data []byte) (string, error) {
	if p.topic == nil {
		return "", fmt.Errorf("topic is nil")
	}

	faster := context.Background()
	r := p.topic.Publish(faster, &pubsub.Message{
		Data: data,
	})
	return r.Get(faster)
}

func (p *Pub) Subscription(ctx context.Context, name string) (*pubsub.Subscription, error) {
	if p.topic == nil {
		return nil, fmt.Errorf("topic is nil")
	}

	return getOrCreateSubscription(ctx, p.cli, p.topic, name)
}

func (p *Pub) Close() error {
	p.topic.Stop()
	return p.cli.Close()
}

func Open(cfg *Config) (*Pub, error) {
	ctx := context.Background()
	cli, err := pubsub.NewClient(ctx, cfg.ProjectId)
	if err != nil {
		return nil, err
	}

	topic, err := getOrCreateTopic(ctx, cli, cfg.TopicId)
	if err != nil {
		return nil, err
	}

	return &Pub{
		cli:   cli,
		topic: topic,
	}, nil
}
