package g

import (
	"context"

	"cloud.google.com/go/pubsub"
	"google.golang.org/api/iterator"
)

func Reset(opts ...OptionFunc) error {
	o := NewOptions()
	for _, opt := range opts {
		opt(o)
	}

	ctx := context.Background()
	cli, err := pubsub.NewClient(ctx, o.ProjectId)
	if err != nil {
		return err
	}

	topic := cli.Topic(o.TopicId)
	ok, err := topic.Exists(ctx)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	subs := topic.Subscriptions(ctx)

	// delete all subscriptions.
	for {
		sub, err := subs.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		if err := sub.Delete(ctx); err != nil {
			return err
		}
	}

	return topic.Delete(ctx)
}
