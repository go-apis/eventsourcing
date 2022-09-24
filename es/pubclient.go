package es

import (
	"context"
)

type SubscriptionCallback func(ctx context.Context, data []byte) error

type Subscription interface {
	Receive(ctx context.Context, callback SubscriptionCallback) error
}

type PubClient interface {
	Publish(ctx context.Context, data []byte) (string, error)
	Subscription(ctx context.Context, suffix string) (Subscription, error)
	Close() error
}
