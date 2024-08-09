package aggregates

import (
	"context"

	"github.com/go-apis/eventsourcing/examples/groups/data/commands"
	"github.com/go-apis/eventsourcing/examples/groups/data/events"
	"github.com/go-apis/eventsourcing/examples/groups/models"
	"github.com/google/uuid"

	"github.com/go-apis/eventsourcing/es"
	"github.com/go-apis/eventsourcing/es/types"
)

type Group struct {
	es.BaseAggregateSourced `es:"group,snapshot=1,rev=rev3,project=false"`

	Name     string
	ParentId *uuid.UUID
	Users    types.Slice[models.User] `gorm:"type:jsonb;serializer:json"`
}

func (u *Group) HandleCreate(ctx context.Context, cmd *commands.CreateGroup) error {
	return u.Apply(ctx, &events.GroupCreated{
		Name:     cmd.Name,
		ParentId: cmd.ParentId,
	})
}
func (u *Group) HandleAddUser(ctx context.Context, cmd *commands.AddUser) error {
	return u.Apply(ctx, &events.UserAdded{
		Users: types.SliceItem[models.User]{
			Index: len(u.Users),
			Value: models.User{
				Id: cmd.UserId,
			},
		},
	})
}
