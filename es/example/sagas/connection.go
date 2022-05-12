package sagas

import (
	"context"
	"eventstore/es"
	"eventstore/es/example/commands"
	"eventstore/es/example/events"
)

type ConnectionSaga struct {
	es.Dispatcher
}

func (s *ConnectionSaga) HandleConnectionAdded(ctx context.Context, evt es.Event, data *events.ConnectionAdded) error {
	item := data.Connections.Value

	cmds := []es.Command{
		&commands.CreateExternalUser{
			Name:     item.Name,
			UserId:   item.UserId,
			Username: item.Username,
		},
	}

	return s.Dispatch(ctx, cmds...)
}

func NewConnectionSaga(dispatcher es.Dispatcher) *ConnectionSaga {
	return &ConnectionSaga{
		Dispatcher: dispatcher,
	}
}
