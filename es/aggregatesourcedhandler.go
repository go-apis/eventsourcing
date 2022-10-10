package es

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
)

type sourcedAggregateHandler struct {
	name    string
	handles CommandHandles
}

func (b *sourcedAggregateHandler) inner(ctx context.Context, entity Entity, cmd Command) error {
	switch agg := entity.(type) {
	case CommandHandler:
		return agg.Handle(ctx, cmd)
	}

	if b.handles != nil {
		return b.handles.Handle(entity, ctx, cmd)
	}

	return fmt.Errorf("no handler for command: %T", cmd)
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
		if err := b.inner(pctx, agg, cmd); err != nil {
			return err
		}

		// what about owner
		// what about parent
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
