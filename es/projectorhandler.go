package es

import (
	"context"

	"go.opentelemetry.io/otel"
)

type projectorEventHandler struct {
	cfg       *EntityConfig
	handles   []*ProjectorHandle
	projector IsProjector
}

func (p *projectorEventHandler) Handle(ctx context.Context, evt *Event) error {
	pctx, pspan := otel.Tracer("projectorEventHandler").Start(ctx, "Handle")
	defer pspan.End()

	unit, err := GetUnit(pctx)
	if err != nil {
		return err
	}

	for _, h := range p.handles {
		// load up the type!.
		ent, err := unit.Load(pctx, p.cfg.Name, evt.AggregateId)
		if err != nil {
			return err
		}

		if err := h.Handle(p.projector, pctx, ent, evt); err != nil {
			return err
		}

		// save it!
		if err := unit.Save(pctx, p.cfg.Name, ent); err != nil {
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
