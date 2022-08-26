package es

import (
	"context"

	"go.opentelemetry.io/otel"
)

type sourcedAggregateHandler struct {
	name    string
	handles CommandHandles
}

func (b *sourcedAggregateHandler) Handle(ctx context.Context, cmd Command) error {
	pctx, pspan := otel.Tracer("SourcedAggregateHandler").Start(ctx, "Handle")
	defer pspan.End()

	unit, err := GetUnit(pctx)
	if err != nil {
		return err
	}
	aggregateId := cmd.GetAggregateId()
	replay := IsReplayCommand(cmd)

	agg, err := unit.Load(pctx, b.name, aggregateId)
	if err != nil {
		return err
	}

	if !replay {
		if err := b.handles.Handle(agg, pctx, cmd); err != nil {
			return err
		}
	}

	if err := unit.Save(pctx, b.name, agg); err != nil {
		return err
	}
	return nil
}

func NewSourcedAggregateHandler(name string, handles CommandHandles) CommandHandler {
	return &sourcedAggregateHandler{
		name:    name,
		handles: handles,
	}
}
