package es

import (
	"context"
)

type TypeBuilder = func() (interface{}, error)

type Callback func(ctx context.Context, evt *Event) error

type Streamer interface {
	Start(ctx context.Context, opts InitializeOptions, callback Callback) error
	Publish(ctx context.Context, evt ...*Event) error
	Close(ctx context.Context) error
}
