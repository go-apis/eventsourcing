package pg

import (
	"context"

	"go.opentelemetry.io/otel"
	"gorm.io/gorm"
)

type RollbackFunc func() error
type CommitFunc func() error

type transaction struct {
	db *gorm.DB

	commitFunc   CommitFunc
	rollbackFunc RollbackFunc
}

func (t *transaction) Commit(ctx context.Context) error {
	_, span := otel.Tracer("local").Start(ctx, "Commit")
	defer span.End()

	if t.commitFunc != nil {
		return t.commitFunc()
	}
	return t.db.Commit().Error
}
func (t *transaction) Rollback(ctx context.Context) error {
	_, span := otel.Tracer("local").Start(ctx, "Rollback")
	defer span.End()

	if t.rollbackFunc != nil {
		return t.rollbackFunc()
	}
	return t.db.Rollback().Error
}
