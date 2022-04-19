package tests

import (
	"context"
	"eventstore/example/aggregates"
	"eventstore/example/commands"
	"eventstore/example/es"
	"log"
	"testing"
)

func Test_It(t *testing.T) {
	// new eventstore
	eventstore := es.NewEventStore(
		es.WithEntities(&aggregates.User{}),
	)

	// connect to eventstore
	client, err := eventstore.Connect()
	if err != nil {
		t.Error(err)
		return
	}

	ctx := context.Background()

	// send a command to store.
	if err := client.Dispatch(ctx, &commands.CreateUser{}); err != nil {
		t.Error(err)
		return
	}

	query := es.NewQuery(client, &aggregates.User{})

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
