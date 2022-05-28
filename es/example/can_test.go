package example

import (
	"context"
	"eventstore/es"
	"eventstore/es/example/aggregates"
	"eventstore/es/example/commands"
	"eventstore/es/example/sagas"
	"eventstore/es/local/pg"
	"testing"
)

func Test_It(t *testing.T) {
	data, err := pg.NewData("postgresql://es:es@localhost:5432/eventstore?sslmode=disable")
	if err != nil {
		t.Error(err)
		return
	}
	// data, err := dg.NewData("localhost:9180")
	// if err != nil {
	// 	t.Error(err)
	// 	return
	// }
	eventstore := es.NewEventStore(data, "example")
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
	ctx, tx, err := data.WithTx(context.Background())
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
	if err := tx.Commit(ctx); err != nil {
		t.Error(err)
		return
	}
}
