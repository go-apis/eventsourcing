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

	if err := unit.Dispatch(pctx, cmds...); err != nil {
		return err
	}

	return nil
}
