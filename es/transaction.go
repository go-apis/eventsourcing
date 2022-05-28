package es

import (
	"context"
	"fmt"
)

var ErrNoTransaction = fmt.Errorf("no transaction")

type txKey int

const thisTxKey txKey = 0

func TransactionCtx(ctx context.Context) (Tx, error) {
	tx, ok := ctx.Value(thisTxKey).(Tx)
	if !ok {
		return nil, ErrNoTransaction
	}
	return tx, nil
}

func SetTransaction(ctx context.Context, tx Tx) context.Context {
	return context.WithValue(ctx, thisTxKey, tx)
}

type Tx interface {
	Load(ctx context.Context, id string, typeName string, out interface{}) error
	Save(ctx context.Context, id string, typeName string, out interface{}) ([]Event, error)
	Commit() error
}
