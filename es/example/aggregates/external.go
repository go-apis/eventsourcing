package aggregates

import (
	"context"
	"eventstore/es"
	"eventstore/es/example/commands"
	"eventstore/es/example/events"
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
