package es

import (
	"context"
)

type Data interface {
	Initialize(cfg Config) error
	NewTx(ctx context.Context) (Tx, error)
}
