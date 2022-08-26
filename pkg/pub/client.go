package pub

import (
	"context"
	"fmt"

	"cloud.google.com/go/pubsub"
)

type Callback func(ctx context.Context, data []byte) error

type Subscription interface {
	Receive(ctx context.Context, callback Callback) error
}

type subscription struct {
	sub *pubsub.Subscription
}

func newSubscription(sub *pubsub.Subscription) *subscription {
	return &subscription{
		sub: sub,
	}
}

func (s *subscription) Receive(ctx context.Context, callback Callback) error {
	return s.sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		if err := callback(ctx, msg.Data); err != nil {
			msg.Nack()
			return
		}
		msg.Ack()
	})
}

type Client interface {
	Publish(ctx context.Context, data []byte) (string, error)
	Subscription(ctx context.Context, suffix string) (Subscription, error)
	Close() error
}

type client struct {
	opts  *Options
	cli   *pubsub.Client
	topic *pubsub.Topic
}

func (c *client) Subscription(ctx context.Context, suffix string) (Subscription, error) {
	if c.topic == nil {
		return nil, fmt.Errorf("topic is nil")
	}

	sub, err := getOrCreateSubscription(ctx, c.cli, c.topic, suffix)
	if err != nil {
		return nil, err
	}

	return newSubscription(sub), nil
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
