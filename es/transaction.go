package es

import (
	"context"
	"fmt"
)

type txKey int

const thisTxKey txKey = 0

var ErrNoTransaction = fmt.Errorf("no transaction")

type TxCommit func() error

type Tx interface {
	Store
	Context() context.Context
	Commit() error
}

type tx struct {
	Store
	ctx    context.Context
	commit TxCommit
}

func (t *tx) Context() context.Context {
	return t.ctx
}
func (t *tx) Commit() error {
	return t.commit()
}

func NewTx(ctx context.Context, store Store, fnCommit TxCommit) (Tx, error) {
	t := &tx{
		Store:  store,
		commit: fnCommit,
	}
	nCtx := context.WithValue(ctx, thisTxKey, t)
	t.ctx = nCtx

	return t, nil
}

func TransactionCtx(ctx context.Context) (Tx, error) {
	tx, ok := ctx.Value(thisTxKey).(Tx)
	if !ok {
		return nil, ErrNoTransaction
	}
	return tx, nil
}
