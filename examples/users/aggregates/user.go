package aggregates

import (
	"context"
	"fmt"

	"github.com/contextcloud/eventstore/es/types"
	"github.com/contextcloud/eventstore/examples/users/commands"
	"github.com/contextcloud/eventstore/examples/users/events"
	"github.com/contextcloud/eventstore/examples/users/models"

	"github.com/contextcloud/eventstore/es"
)

type User struct {
	es.BaseAggregateSourced

	Username    string
	Password    string
	Email       string
	Connections types.Slice[models.Connection] `gorm:"type:jsonb;serializer:json"`
	Groups      types.Slice[models.Group]      `gorm:"type:jsonb;serializer:json"`
}

func (u *User) HandleCreate(ctx context.Context, cmd *commands.CreateUser) error {
	return u.Apply(ctx, events.UserCreated{
		Username: cmd.Username,
		Password: cmd.Password,
	})
}
func (u *User) HandleAddEmail(ctx context.Context, cmd *commands.AddEmail) error {
	return u.Apply(ctx, events.EmailAdded{
		Email: cmd.Email,
	})
}

func (u *User) HandleAddConnection(ctx context.Context, cmd *commands.AddConnection) error {
	return u.Apply(ctx, events.ConnectionAdded{
		Connections: types.SliceItem[models.Connection]{
			Index: len(u.Connections),
			Value: models.Connection{
				Name:     cmd.Name,
				UserId:   cmd.UserId,
				Username: cmd.Username,
			},
		},
	})
}

func (u *User) HandleAddGroup(ctx context.Context, cmd *commands.AddGroup) error {
	return u.Apply(ctx, events.GroupAdded{
		Groups: types.SliceItem[models.Group]{
			Index: len(u.Groups),
			Value: models.Group{
				Id:   cmd.GroupId,
				Name: cmd.Name,
			},
		},
	})
}

func (u *User) HandleUpdateConnection(ctx context.Context, cmd *commands.UpdateConnection) error {
	if len(u.Connections) == 0 {
		return fmt.Errorf("Can't update connection")
	}

	return u.Apply(ctx, events.ConnectionUpdated{
		Connections: types.SliceItem[models.ConnectionUpdate]{
			Index: 0,
			Value: models.ConnectionUpdate{
				Username: cmd.Username,
			},
		},
	})
}
