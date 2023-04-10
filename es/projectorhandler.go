package es

import (
	"context"
)

type projectorEventHandler struct {
	cfg       *EntityConfig
	handles   []*ProjectorHandle
	projector IsProjector
}

func (p *projectorEventHandler) Handle(ctx context.Context, evt *Event) error {
	unit, err := GetUnit(ctx)
	if err != nil {
		return err
	}

	for _, h := range p.handles {
		// load up the type!.
		ent, err := unit.Load(ctx, p.cfg.Name, evt.AggregateId)
		if err != nil {
			return err
		}

		if err := h.Handle(p.projector, ctx, ent, evt); err != nil {
			return err
		}

		// save it!
		if err := unit.Save(ctx, p.cfg.Name, ent); err != nil {
			return err
		}
	}

	return nil
}

func NewProjectorEventHandler(cfg *EntityConfig, handles []*ProjectorHandle, projector IsProjector) EventHandler {
	return &projectorEventHandler{
		cfg:       cfg,
		handles:   handles,
		projector: projector,
	}
}
