package aggregates

import (
	"context"
	"eventstore/es"
	"eventstore/es/example/commands"
	"eventstore/es/example/events"
	"eventstore/es/types"
)

type User struct {
	es.BaseAggregate

	Username string
	Password types.Encrypted
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
