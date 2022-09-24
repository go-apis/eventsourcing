package gpub

import (
	"context"

	"cloud.google.com/go/pubsub"
	"google.golang.org/api/iterator"
)

func Reset(cfg *Config) error {
	ctx := context.Background()
	cli, err := pubsub.NewClient(ctx, cfg.ProjectId)
	if err != nil {
		return err
	}

	topic := cli.Topic(cfg.TopicId)
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
