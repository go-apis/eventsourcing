package sagas

import (
	"context"

	"github.com/contextcloud/eventstore/examples/groups/commands"
	"github.com/contextcloud/eventstore/examples/groups/events"

	"github.com/contextcloud/eventstore/es"
)

type UserSaga struct {
	es.BaseSaga
}

func (s *UserSaga) HandleConnectionAdded(ctx context.Context, evt *es.Event, data *events.GroupAdded) ([]es.Command, error) {
	return es.Commands(&commands.AddUser{
		BaseCommand: es.BaseCommand{
			AggregateId: data.Groups.Value.Id,
		},
		UserId: evt.AggregateId,
	}), nil
}

func NewUserSaga() *UserSaga {
	return &UserSaga{}
}
