package aggregates

import (
	"context"
	"eventstore/example/commands"
	"eventstore/example/es"
	"eventstore/example/events"
)

type User struct {
	es.BaseAggregate
}

func (u *User) HandleCreate(ctx context.Context, cmd *commands.CreateUser) error {
	u.Apply(ctx, events.UserCreated{
		Username: cmd.Username,
		Password: cmd.Password,
	})
	return nil
}
