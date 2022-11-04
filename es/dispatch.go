package es

import (
	"context"

	"go.opentelemetry.io/otel"
)

func Dispatch(ctx context.Context, cmds ...Command) error {
	pctx, span := otel.Tracer("es").Start(ctx, "Dispatch")
	defer span.End()

	unit, err := GetUnit(pctx)
	if err != nil {
		return err
	}

	tx, err := unit.NewTx(pctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(pctx)

	if err := unit.Dispatch(pctx, cmds...); err != nil {
		return err
	}

	if _, err := tx.Commit(pctx); err != nil {
		return err
	}

	return nil
}
