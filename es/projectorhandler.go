package es

import (
	"context"
	"errors"
)

var ErrDeleteEntity = errors.New("delete entity")

type projectorEventHandler struct {
	cfg       *EntityConfig
	handles   []*ProjectorHandle
	projector IsProjector
}

func (p *projectorEventHandler) HandleEvent(ctx context.Context, evt *Event) error {
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

		if err := h.Handle(p.projector, ctx, ent, evt); err != nil && err != ErrDeleteEntity {
			return err
		} else if err == ErrDeleteEntity {
			// delete it.
			if err := unit.Delete(ctx, p.cfg.Name, ent); err != nil {
				return err
			}
		} else {
			// save it!
			if err := unit.Save(ctx, p.cfg.Name, ent); err != nil {
				return err
			}
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
