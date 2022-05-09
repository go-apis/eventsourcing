package aggregates

import (
	"context"
	"eventstore/es"
	"eventstore/es/example/commands"
	"eventstore/es/example/events"
	"eventstore/es/example/models"
	"eventstore/es/types"
	"fmt"
)

type User struct {
	es.BaseAggregate

	Username    string
	Password    types.Encrypted
	Email       string
	Connections types.Slice[models.Connection]
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
