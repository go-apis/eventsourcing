package sagas

import (
	"context"

	"github.com/contextcloud/eventstore/examples/users/commands"
	"github.com/contextcloud/eventstore/examples/users/events"

	"github.com/contextcloud/eventstore/es"
)

type GroupSaga struct {
	es.BaseSaga
}

func (s *GroupSaga) HandleConnectionAdded(ctx context.Context, evt es.Event, data events.UserAdded) ([]es.Command, error) {
	return es.Commands(&commands.AddGroup{
		BaseCommand: es.BaseCommand{
			AggregateId: evt.AggregateId,
		},
		GroupId: evt.AggregateId,
		Name:    data.Name,
	}), nil
}

func NewGroupSaga() *GroupSaga {
	return &GroupSaga{}
}
