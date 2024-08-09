package handlers

import (
	"context"

	"github.com/go-apis/eventsourcing/es"
	"github.com/go-apis/eventsourcing/examples/groups/data/aggregates"
	"github.com/go-apis/eventsourcing/examples/groups/data/commands"
	"github.com/go-apis/eventsourcing/examples/groups/data/events"
)

type demoHandler struct {
	es.BaseCommandHandler

	s es.Store[*aggregates.Demo]
}

func (h *demoHandler) Handle(ctx context.Context, cmd *commands.NewDemo) error {
	d, err := h.s.Load(ctx, cmd.AggregateId)
	if err != nil {
		return err
	}

	d.Name = cmd.Name
	d.Publish(&events.DemoCreated{
		Name: cmd.Name,
	})

	return h.s.Save(ctx, d)
}

func NewDemoHandler() es.IsCommandHandler {
	return &demoHandler{
		s: es.NewStore[*aggregates.Demo](),
	}
}
