package aggregates

import (
	"context"

	"github.com/go-apis/eventsourcing/examples/users/data/commands"
	"github.com/go-apis/eventsourcing/examples/users/data/events"

	"github.com/go-apis/eventsourcing/es"
)

type ExternalUser struct {
	es.BaseAggregateSourced `es:"external_user,snapshot=3"`

	Name     string
	UserId   string
	Username string
}

func (u *ExternalUser) HandleCreate(ctx context.Context, cmd *commands.CreateExternalUser) error {
	return u.Apply(ctx, &events.ExternalUserCreated{
		Name:     cmd.Name,
		UserId:   cmd.UserId,
		Username: cmd.Username,
	})
}
