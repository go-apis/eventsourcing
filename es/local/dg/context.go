package dg

import (
	"context"
	"fmt"
)

var ErrNoTransaction = fmt.Errorf("no transaction")

type txKey int

const thisTxKey txKey = 0

func transactionCtx(ctx context.Context) (*tx, error) {
	tx, ok := ctx.Value(thisTxKey).(*tx)
	if !ok {
		return nil, ErrNoTransaction
	}
	return tx, nil
}

func setTransaction(ctx context.Context, tx *tx) context.Context {
	return context.WithValue(ctx, thisTxKey, tx)
}
