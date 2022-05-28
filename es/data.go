package es

import "context"

type Data interface {
	WithTx(ctx context.Context) (context.Context, Tx, error)
	GetTx(ctx context.Context) (Tx, error)

	GetEvents(ctx context.Context) ([]Event, error)
}
