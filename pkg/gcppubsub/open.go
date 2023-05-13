package gcppubsub

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/pubsub"
)

type Message struct {
	Data     []byte
	OrderKey string
}

type Pub struct {
	cli   *pubsub.Client
	topic *pubsub.Topic
}

func (p *Pub) Publish(ctx context.Context, messages ...*pubsub.Message) ([]string, error) {
	if p.topic == nil {
		return nil, fmt.Errorf("topic is nil")
	}

	var results []*pubsub.PublishResult
	for _, message := range messages {
		result := p.topic.Publish(ctx, message)
		results = append(results, result)
	}

	var resultErrors []error
	var ids []string
	for _, res := range results {
		id, err := res.Get(ctx)
		if err != nil {
			resultErrors = append(resultErrors, err)
			continue
		}
		ids = append(ids, id)
	}

	if len(resultErrors) != 0 {
		return nil, fmt.Errorf("Get: %v", resultErrors[len(resultErrors)-1])
	}
	return ids, nil
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
	topic.EnableMessageOrdering = true
	topic.PublishSettings.ByteThreshold = 5000
	topic.PublishSettings.CountThreshold = 10
	topic.PublishSettings.DelayThreshold = 100 * time.Millisecond

	return &Pub{
		cli:   cli,
		topic: topic,
	}, nil
}
