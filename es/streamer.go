package es

import (
	"context"
)

type EventCallback func(ctx context.Context, evt *Event) error

type Streamer interface {
	Start(ctx context.Context, opts InitializeOptions, callback EventCallback) error
	Publish(ctx context.Context, evt ...*Event) error
	Close(ctx context.Context) error
}
