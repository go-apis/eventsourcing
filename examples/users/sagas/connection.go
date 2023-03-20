package sagas

import (
	"context"

	"github.com/contextcloud/eventstore/examples/users/commands"
	"github.com/contextcloud/eventstore/examples/users/events"
	"github.com/google/uuid"

	"github.com/contextcloud/eventstore/es"
)

type ConnectionSaga struct {
	es.BaseSaga
}

func (s *ConnectionSaga) HandleConnectionAdded(ctx context.Context, evt *es.Event, data *events.ConnectionAdded) ([]es.Command, error) {
	item := data.Connections.Value

	id := uuid.NewSHA1(evt.AggregateId, []byte(item.UserId))

	return es.Commands(&commands.CreateExternalUser{
		BaseCommand: es.BaseCommand{
			AggregateId: id,
		},

		Name:     item.Name,
		UserId:   item.UserId,
		Username: item.Username,
	}), nil
}

func NewConnectionSaga() *ConnectionSaga {
	return &ConnectionSaga{
		BaseSaga: es.BaseSaga{
			IsAsync: true,
		},
	}
}
