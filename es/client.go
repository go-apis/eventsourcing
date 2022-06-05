package es

import "context"

type Client interface {
	NewTx(ctx context.Context) (Tx, error)
	NewSourcedStore(dispatcher Dispatcher, name string) SourcedStore
}
