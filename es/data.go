package es

import (
	"context"
)

type Data interface {
	NewTx(ctx context.Context) (Tx, error)
}
