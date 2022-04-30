package example

import (
	"context"
	"eventstore/es"
	"eventstore/es/example/aggregates"
	"eventstore/es/example/commands"
	"log"
	"testing"
)

func Test_It(t *testing.T) {
	ctx := context.Background()
	cmds := []es.Command{
		&commands.CreateUser{
			BaseCommand: es.BaseCommand{
				AggregateId: "98f1f7d3-f312-4d57-8847-5b9ac8d5797d",
			},
			Username: "chris.kolenko",
			Password: "12345678",
		},
		&commands.AddEmail{
			BaseCommand: es.BaseCommand{
				AggregateId: "98f1f7d3-f312-4d57-8847-5b9ac8d5797d",
			},
			Email: "chris@context.gg",
		},
	}

	// new eventstore
	eventstore, err := es.NewEventStore(
		es.WithDb("postgresql://es:es@localhost:5432/eventstore?sslmode=disable"),
		es.WithHandlers(
			&aggregates.User{},
		),
	)
	if err != nil {
		t.Error(err)
		return
	}

	// TODO: start it?

	for _, cmd := range cmds {
		// send a command to store.
		if err := eventstore.Dispatch(ctx, cmd); err != nil {
			t.Error(err)
			return
		}
	}

	query := es.NewQuery[aggregates.User](eventstore)
	user, err := query.Get(ctx, "98f1f7d3-f312-4d57-8847-5b9ac8d5797d")
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
