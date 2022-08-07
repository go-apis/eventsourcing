package aggregates

import (
	"context"

	"github.com/contextcloud/eventstore/demo/users/commands"
	"github.com/contextcloud/eventstore/demo/users/events"

	"github.com/contextcloud/eventstore/es"
)

type ExternalUser struct {
	es.BaseAggregate `es:"external_user,snapshot=3"`

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
