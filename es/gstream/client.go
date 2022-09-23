package gstream

import (
	"context"
)

type Callback func(ctx context.Context, data []byte) error

type Subscription interface {
	Receive(ctx context.Context, callback Callback) error
}

type Client interface {
	Publish(ctx context.Context, data []byte) (string, error)
	Subscription(ctx context.Context, suffix string) (Subscription, error)
	Close() error
}
