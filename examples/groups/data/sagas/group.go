package sagas

import (
	"context"

	"github.com/go-apis/eventsourcing/examples/groups/data/commands"
	"github.com/go-apis/eventsourcing/examples/groups/data/events"

	"github.com/go-apis/eventsourcing/es"
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
