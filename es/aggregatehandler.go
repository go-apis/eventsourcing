package es

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type aggregateHandler struct {
	cfg     *EntityConfig
	handles CommandHandles
}

func (b *aggregateHandler) inner(ctx context.Context, entity Entity, cmd Command) error {
	switch agg := entity.(type) {
	case CommandHandler:
		return agg.Handle(ctx, cmd)
	}
	if b.handles != nil {
		return b.handles.Handle(entity, ctx, cmd)
	}
	return fmt.Errorf("no handler for command: %T", cmd)
}

func (b *aggregateHandler) Handle(ctx context.Context, cmd Command) error {
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
	}

	if err := unit.Save(pctx, b.cfg.Name, agg); err != nil {
		return err
	}
	return nil
}

func NewAggregateHandler(cfg *EntityConfig, handles CommandHandles) CommandHandler {
	return &aggregateHandler{
		cfg:     cfg,
		handles: handles,
	}
}
