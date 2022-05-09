package example

import (
	"context"
	"eventstore/es"
	"eventstore/es/example/aggregates"
	"eventstore/es/example/commands"
	"testing"
)

func Test_It(t *testing.T) {
	// new eventstore
	eventstore, err := es.NewEventStore(
		es.WithDb("postgresql://es:es@localhost:5432/eventstore?sslmode=disable"),
		es.WithServiceName("example"),
		es.WithHandlers(
			&aggregates.User{},
		),
	)
	if err != nil {
		t.Error(err)
		return
	}

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
		&commands.AddConnection{
			BaseCommand: es.BaseCommand{
				AggregateId: "98f1f7d3-f312-4d57-8847-5b9ac8d5797d",
			},
			Name:     "Smashgg",
			UserId:   "demo1",
			Username: "chris.kolenko",
		},
		&commands.UpdateConnection{
			BaseCommand: es.BaseCommand{
				AggregateId: "98f1f7d3-f312-4d57-8847-5b9ac8d5797d",
			},
			Username: "aaaaaaaaaa",
		},
		&commands.Fail{
			BaseCommand: es.BaseCommand{
				AggregateId: "98f1f7d3-f312-4d57-8847-5b9ac8d5797d",
			},
		},
	}

	// the event store should know the aggregates and the commands.
	tx := eventstore.Get(ctx)

	for _, cmd := range cmds {
		// send a command to store.
		if err := tx.Dispatch(ctx, cmd); err != nil {
			t.Error(err)
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		t.Error(err)
		return
	}
}
