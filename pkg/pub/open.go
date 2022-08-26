package pub

import (
	"context"

	"cloud.google.com/go/pubsub"
)

func Open(opts ...OptionFunc) (Client, error) {
	o := NewOptions()
	for _, opt := range opts {
		opt(o)
	}

	ctx := context.Background()
	cli, err := pubsub.NewClient(ctx, o.ProjectId)
	if err != nil {
		return nil, err
	}

	topic, err := getOrCreateTopic(ctx, cli, o.TopicId)
	if err != nil {
		return nil, err
	}

	return &client{
		opts:  o,
		cli:   cli,
		topic: topic,
	}, nil
}
