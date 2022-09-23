package g

import (
	"context"
	"fmt"

	"cloud.google.com/go/pubsub"
	"github.com/contextcloud/eventstore/es/gstream"
)

type client struct {
	opts  *Options
	cli   *pubsub.Client
	topic *pubsub.Topic
}

func (c *client) Subscription(ctx context.Context, suffix string) (gstream.Subscription, error) {
	if c.topic == nil {
		return nil, fmt.Errorf("topic is nil")
	}

	sub, err := getOrCreateSubscription(ctx, c.cli, c.topic, suffix)
	if err != nil {
		return nil, err
	}

	return NewSubscription(sub), nil
}

func (c *client) Publish(ctx context.Context, data []byte) (string, error) {
	if c.topic == nil {
		return "", fmt.Errorf("topic is nil")
	}

	faster := context.Background()
	r := c.topic.Publish(faster, &pubsub.Message{
		Data: data,
	})
	return r.Get(faster)
}

func (c *client) Close() error {
	c.topic.Stop()
	return c.cli.Close()
}
