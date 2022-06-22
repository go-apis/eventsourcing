package aggregates

import (
	"context"

	"github.com/contextcloud/eventstore/es/example/commands"
	"github.com/contextcloud/eventstore/es/example/events"

	"github.com/contextcloud/eventstore/es"
)

type ExternalUser struct {
	es.BaseAggregate

	Name     string
	UserId   string
	Username string
}

func (u *ExternalUser) HandleCreate(ctx context.Context, cmd *commands.CreateExternalUser) error {
	return u.Apply(ctx, events.ExternalUserCreated{
		Name:     cmd.Name,
		UserId:   cmd.UserId,
		Username: cmd.Username,
	})
}
