package handlers

import (
	"context"

	"github.com/contextcloud/eventstore/es"
	"github.com/contextcloud/eventstore/examples/groups/data/aggregates"
	"github.com/contextcloud/eventstore/examples/groups/data/commands"
	"github.com/contextcloud/eventstore/examples/groups/data/events"
)

type demoHandler struct {
	es.BaseCommmandHandler

	q es.Query[*aggregates.Demo]
}

func (h *demoHandler) Handle(ctx context.Context, cmd *commands.NewDemo) error {
	d, err := h.q.Load(ctx, cmd.AggregateId)
	if err != nil {
		return err
	}

	d.Name = cmd.Name
	d.Publish(&events.DemoCreated{
		Name: cmd.Name,
	})

	return h.q.Save(ctx, d)
}

func NewDemoHandler() es.IsCommandHandler {
	return &demoHandler{
		q: es.NewQuery[*aggregates.Demo](),
	}
}
