package tests

import (
	"context"
	"eventstore/es"
	"eventstore/es/example/aggregates"
	"eventstore/es/example/commands"
	"log"
	"testing"
)

func Test_It(t *testing.T) {
	// new eventstore
	eventstore, err := es.NewEventStore(
		es.WithDb(),
		es.WithHandlers(
			&aggregates.User{},
		),
	)

	// TODO: start it?

	ctx := context.Background()

	// send a command to store.
	if err := eventstore.Dispatch(ctx, &commands.CreateUser{}); err != nil {
		t.Error(err)
		return
	}

	query := es.NewQuery(eventstore, &aggregates.User{})
	user, err := query.Find(ctx, "user-1")
	if err != nil {
		t.Error(err)
		return
	}

	if user == nil {
		t.Error("Invalid user")
		return
	}

	log.Printf("user: %+v", user)
}
