package es

import "context"

type SourcedStore interface {
	WithTx(ctx context.Context) (context.Context, error)
	GetTx(ctx context.Context) (Tx, error)
}

type sourceStore struct {
}

func NewSourcedStore(data Data) SourcedStore {
	return &sourceStore{}
}
