package es

import "context"

type Client interface {
	WithTx(ctx context.Context) (context.Context, Tx, error)
	GetTx(ctx context.Context) (Tx, error)
	NewSourcedStore(dispatcher Dispatcher, name string) SourcedStore
}
