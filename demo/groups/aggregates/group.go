package aggregates

import (
	"context"

	"github.com/contextcloud/eventstore/demo/groups/commands"
	"github.com/contextcloud/eventstore/demo/groups/events"
	"github.com/google/uuid"

	"github.com/contextcloud/eventstore/es"
)

type Group struct {
	es.BaseAggregate

	Name     string
	ParentId *uuid.UUID
}

func (u *Group) HandleCreate(ctx context.Context, cmd *commands.CreateGroup) error {
	return u.Apply(ctx, events.GroupCreated{
		Name:     cmd.Name,
		ParentId: cmd.ParentId,
	})
}
func (u *Group) HandleAddUser(ctx context.Context, cmd *commands.AddUser) error {
	return u.Apply(ctx, events.UserAdded{
		UserId: cmd.UserId,
	})
}
