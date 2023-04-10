package es

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type sourcedAggregateHandler struct {
	cfg     *EntityConfig
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

	pspan.SetAttributes(
		attribute.String("aggregate", b.cfg.Name),
		attribute.String("id", aggregateId.String()),
		attribute.Bool("replay", replay),
	)

	agg, err := unit.Load(pctx, b.cfg.Name, aggregateId)
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

	if err := unit.Save(pctx, b.cfg.Name, agg); err != nil {
		return err
	}
	return nil
}

func NewSourcedAggregateHandler(cfg *EntityConfig, handles CommandHandles) CommandHandler {
	return &sourcedAggregateHandler{
		cfg:     cfg,
		handles: handles,
	}
}
