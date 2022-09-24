package gpub

import (
	"context"
	"time"

	"cloud.google.com/go/pubsub"
)

func getOrCreateTopic(ctx context.Context, cli *pubsub.Client, topicName string) (*pubsub.Topic, error) {
	topic := cli.Topic(topicName)

	ok, err := topic.Exists(ctx)
	if err != nil {
		return nil, err
	}
	if ok {
		return topic, nil
	}

	created, err := cli.CreateTopic(ctx, topicName)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func getOrCreateSubscription(ctx context.Context, cli *pubsub.Client, topic *pubsub.Topic, suffix string) (*pubsub.Subscription, error) {
	subscriptionId := topic.ID() + "__" + suffix
	sub := cli.Subscription(subscriptionId)
	ok, err := sub.Exists(ctx)
	if err != nil {
		return nil, err
	}
	if ok {
		return sub, nil
	}

	created, err := cli.CreateSubscription(ctx, subscriptionId, pubsub.SubscriptionConfig{
		Topic:       topic,
		AckDeadline: 60 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}
