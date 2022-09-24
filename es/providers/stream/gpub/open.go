package gpub

import (
	"context"

	"cloud.google.com/go/pubsub"
	"github.com/contextcloud/eventstore/pkg/pub"
)

func Open(cfg *Config) (pub.Client, error) {
	ctx := context.Background()
	cli, err := pubsub.NewClient(ctx, cfg.ProjectId)
	if err != nil {
		return nil, err
	}

	topic, err := getOrCreateTopic(ctx, cli, cfg.TopicId)
	if err != nil {
		return nil, err
	}

	return &client{
		cli:   cli,
		topic: topic,
	}, nil
}
