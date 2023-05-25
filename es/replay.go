package es

import (
	"context"

	"go.opentelemetry.io/otel"
)

func Replay(ctx context.Context, cmds ...*ReplayCommand) error {
	pctx, span := otel.Tracer("es").Start(ctx, "Replay")
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

	if err := unit.Replay(pctx, cmds...); err != nil {
		return err
	}

	if _, err := tx.Commit(pctx); err != nil {
		return err
	}

	return nil
}
