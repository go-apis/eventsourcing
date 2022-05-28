package example

import (
	"context"
	"eventstore/es"
	"eventstore/es/example/aggregates"
	"eventstore/es/example/commands"
	"eventstore/es/example/sagas"
	"eventstore/es/local"
	"testing"
)

func Test_It(t *testing.T) {
	factory, err := local.NewPostgresData("postgresql://es:es@localhost:5432/eventstore?sslmode=disable")
	if err != nil {
		t.Error(err)
		return
	}
	eventstore := es.NewEventStore(factory, "example")
	if err != nil {
		t.Error(err)
		return
	}
	eventstore.AddCommandHandler(
		&aggregates.User{},
		&aggregates.ExternalUser{},
	)
	eventstore.AddEventHandler(
		sagas.NewConnectionSaga(eventstore),
	)

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
	ctx, tx, err := factory.WithTx(context.Background())
	if err != nil {
		t.Error(err)
		return
	}

	// send a commands to store.
	if err := eventstore.Dispatch(ctx, cmds...); err != nil {
		t.Error(err)
		return
	}

	// commit the tx.
	if err := tx.Commit(); err != nil {
		t.Error(err)
		return
	}
}
