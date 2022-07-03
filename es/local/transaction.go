package local

import (
	"context"

	"github.com/contextcloud/eventstore/es"
)

type transaction struct {
	d *data
}

func (t *transaction) Commit(ctx context.Context) (int, error) {
	out := t.d.tx.Commit()
	t.d.isCommitted = true
	return int(out.RowsAffected), out.Error
}
func (t *transaction) Rollback(ctx context.Context) error {
	out := t.d.tx.Rollback()
	return out.Error
}

func newTransaction(d *data) es.Tx {
	return &transaction{
		d: d,
	}
}
