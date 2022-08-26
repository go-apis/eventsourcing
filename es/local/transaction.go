package local

import (
	"context"

	"github.com/contextcloud/eventstore/es"
	"go.opentelemetry.io/otel"
)

type transaction struct {
	d *data
}

func (t *transaction) Commit(ctx context.Context) (int, error) {
	_, span := otel.Tracer("local").Start(ctx, "Commit")
	defer span.End()

	out := t.d.tx.Commit()
	t.d.isCommitted = true
	return int(out.RowsAffected), out.Error
}
func (t *transaction) Rollback(ctx context.Context) error {
	_, span := otel.Tracer("local").Start(ctx, "Rollback")
	defer span.End()

	out := t.d.tx.Rollback()
	return out.Error
}

func newTransaction(d *data) es.Tx {
	return &transaction{
		d: d,
	}
}
