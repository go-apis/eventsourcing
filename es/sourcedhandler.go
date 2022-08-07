package es

import (
	"context"
)

type aggregateHandler struct {
	name    string
	handles CommandHandles
}

func (b *aggregateHandler) Handle(ctx context.Context, cmd Command) error {
	unit := UnitFromContext(ctx)
	aggregateId := cmd.GetAggregateId()
	replay := IsReplayCommand(cmd)

	agg, err := unit.Load(ctx, b.name, aggregateId)
	if err != nil {
		return err
	}

	if !replay {
		if err := b.handles.Handle(agg, ctx, cmd); err != nil {
			return err
		}
	}

	if err := unit.Save(ctx, b.name, agg); err != nil {
		return err
	}
	return nil
}

func NewSourcedAggregateHandler(name string, handles CommandHandles) CommandHandler {
	return &aggregateHandler{
		name:    name,
		handles: handles,
	}
}
