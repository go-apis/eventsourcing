package tests

import (
	"context"
	"eventstore/es"
	"eventstore/es/example/aggregates"
	"eventstore/es/example/commands"
	"eventstore/es/local"
	"eventstore/es/local/pg"
	"testing"
)

func Test_It(t *testing.T) {
	data, err := pg.NewData("postgresql://es:es@localhost:5432/eventstore?sslmode=disable")
	if err != nil {
		t.Error(err)
		return
	}
	cli := local.NewClient(data, "example")

	eventstore := es.NewEventStore(cli)
	if err != nil {
		t.Error(err)
		return
	}
	eventstore.AddCommandHandler(
		&aggregates.User{},
		&aggregates.ExternalUser{},
	)
	// eventstore.AddEventHandler(
	// 	sagas.NewConnectionSaga(eventstore),
	// )

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

	ctx := context.Background()

	// the event store should know the aggregates and the commands.
	unit, err := eventstore.NewUnit(ctx)
	if err != nil {
		t.Error(err)
		return
	}

	// send a commands to store.
	if err := unit.Dispatch(ctx, cmds...); err != nil {
		t.Error(err)
		return
	}

	// commit the tx.
	if err := unit.Commit(ctx); err != nil {
		t.Error(err)
		return
	}
}
