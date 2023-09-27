package sagas

import (
	"context"
	"fmt"

	"github.com/contextcloud/eventstore/es/filters"
	"github.com/contextcloud/eventstore/examples/users/data/aggregates"
	"github.com/contextcloud/eventstore/examples/users/data/commands"
	"github.com/contextcloud/eventstore/examples/users/data/events"
	"github.com/google/uuid"

	"github.com/contextcloud/eventstore/es"
)

type ConnectionSaga struct {
	es.BaseSaga
}

func (s *ConnectionSaga) HandleConnectionAdded(ctx context.Context, evt *es.Event, data *events.ConnectionAdded) ([]es.Command, error) {
	item := data.Connections.Value

	id := uuid.NewSHA1(evt.AggregateId, []byte(item.UserId))

	q := es.NewQuery[*aggregates.User]()
	all, err := q.Find(ctx, filters.Filter{})
	if err != nil {
		return nil, err
	}
	fmt.Printf("all: %+v", all)

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
		BaseSaga: es.BaseSaga{},
	}
}
