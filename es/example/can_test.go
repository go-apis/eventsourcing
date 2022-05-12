package example

import (
	"context"
	"eventstore/es"
	"eventstore/es/example/aggregates"
	"eventstore/es/example/commands"
	"eventstore/es/example/sagas"
	"testing"
)

func Test_It(t *testing.T) {
	eventstore, err := es.NewEventStore("example", "postgresql://es:es@localhost:5432/eventstore?sslmode=disable")
	if err != nil {
		t.Error(err)
		return
	}

	if err := eventstore.Config(
		es.WithCommandHandlers(
			&aggregates.User{},
			&aggregates.ExternalUser{},
		),
		es.WithEventHandlers(
			sagas.NewConnectionSaga(eventstore),
		),
	); err != nil {
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
	}

	// the event store should know the aggregates and the commands.
	tx, err := eventstore.NewTx(ctx)
	if err != nil {
		t.Error(err)
		return
	}

	ctx = tx.Context()

	// send a commands to store.
	if err := eventstore.Dispatch(ctx, cmds...); err != nil {
		t.Error(err)
		return
	}

	if err := tx.Commit(); err != nil {
		t.Error(err)
		return
	}
}
