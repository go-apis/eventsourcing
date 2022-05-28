package dg

import (
	"context"
	"eventstore/es"

	"github.com/dgraph-io/dgo/v210"
)

type tx struct {
	d   *data
	txn *dgo.Txn
}

func (t *tx) Commit(ctx context.Context) error {
	return t.txn.Commit(ctx)
}

func newTx(d *data) es.Tx {
	txn := d.cli.NewTxn()

	return &tx{
		d:   d,
		txn: txn,
	}
}
