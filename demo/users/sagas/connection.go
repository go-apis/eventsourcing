package sagas

import (
	"context"

	"github.com/contextcloud/eventstore/demo/users/commands"
	"github.com/contextcloud/eventstore/demo/users/events"

	"github.com/contextcloud/eventstore/es"
)

type ConnectionSaga struct {
	es.BaseSaga
}

func (s *ConnectionSaga) HandleConnectionAdded(ctx context.Context, evt es.Event, data events.ConnectionAdded) ([]es.Command, error) {
	item := data.Connections.Value
	return es.Commands(&commands.CreateExternalUser{
		BaseCommand: es.BaseCommand{
			AggregateId: evt.AggregateId,
		},

		Name:     item.Name,
		UserId:   item.UserId,
		Username: item.Username,
	}), nil
}

func NewConnectionSaga() *ConnectionSaga {
	return &ConnectionSaga{}
}
