package sagas

import (
	"context"

	"github.com/contextcloud/eventstore/es/example/commands"
	"github.com/contextcloud/eventstore/es/example/events"

	"github.com/contextcloud/eventstore/es"
)

type ConnectionSaga struct {
	es.BaseSaga
}

func (s *ConnectionSaga) HandleConnectionAdded(ctx context.Context, evt es.Event, data events.ConnectionAdded) error {
	item := data.Connections.Value
	s.Handle(ctx, &commands.CreateExternalUser{
		BaseCommand: es.BaseCommand{
			AggregateId: evt.AggregateId,
		},

		Name:     item.Name,
		UserId:   item.UserId,
		Username: item.Username,
	})
	return nil
}

func NewConnectionSaga() *ConnectionSaga {
	return &ConnectionSaga{}
}
