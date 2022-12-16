package handlers

import (
	"context"

	"github.com/contextcloud/eventstore/es"
	"github.com/contextcloud/eventstore/examples/groups/data/aggregates"
	"github.com/contextcloud/eventstore/examples/groups/data/commands"
)

type Demo struct {
	es.BaseCommmandHandler

	q es.Query[*aggregates.MyName]
}

func (d *Demo) HandleDemo(ctx context.Context, cmd *commands.DemoCommand) error {
	my, err := d.q.Load(ctx, cmd.AggregateId)
	if err != nil {
		return err
	}

	my.Name = cmd.Name
	return d.q.Save(ctx, my)
}

func NewDemo() *Demo {
	q := es.NewQuery[*aggregates.MyName]()

	return &Demo{
		q: q,
	}
}
