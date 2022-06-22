package tests

import (
	"context"
	"testing"

	"github.com/contextcloud/eventstore/es/example/aggregates"
	"github.com/contextcloud/eventstore/es/example/commands"
	"github.com/contextcloud/eventstore/es/example/sagas"
	"github.com/contextcloud/eventstore/es/filters"
	"github.com/contextcloud/eventstore/es/local/g"

	"github.com/contextcloud/eventstore/es"
)

func Test_It(t *testing.T) {
	cfg, err := es.NewConfig(
		"example",
		&aggregates.User{},
		&aggregates.ExternalUser{},
		sagas.NewConnectionSaga(),
	)
	if err != nil {
		t.Error(err)
		return
	}

	dsn := "postgresql://es:es@localhost:5432/eventstore?sslmode=disable"
	if err := g.ResetDb(dsn); err != nil {
		t.Error(err)
		return
	}

	data, err := g.NewData(cfg, dsn)
	if err != nil {
		t.Error(err)
		return
	}
	cli := es.NewClient(cfg, data)

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
		&commands.CreateUser{
			BaseCommand: es.BaseCommand{
				AggregateId: "2ca16492-ea7a-4d96-8599-b256c26e89b5",
			},
			Username: "calvin.harris",
			Password: "12345678",
		},
	}

	ctx := context.Background()

	// the event store should know the aggregates and the commands.
	unit, err := cli.NewUnit(ctx)
	if err != nil {
		t.Error(err)
		return
	}

	// send a commands to store.
	if err := unit.Dispatch(ctx, cmds...); err != nil {
		t.Error(err)
		return
	}

	userQuery := es.NewQuery[aggregates.User](unit)
	user, err := userQuery.Load(ctx, "98f1f7d3-f312-4d57-8847-5b9ac8d5797d")
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(user)

	filter := filters.Filter{
		Where: filters.WhereClause{
			Column: "username",
			Op:     "eq",
			Args:   []interface{}{"chris.kolenko"},
		},
		Order:  []filters.Order{{Column: "username"}},
		Limit:  filters.Limit(1),
		Offset: filters.Offset(0),
	}

	users, err := userQuery.Find(ctx, filter)
	if err != nil {
		t.Error(err)
		return
	}

	t.Log(users)

	// commit the tx.
	if err := unit.Commit(ctx); err != nil {
		t.Error(err)
		return
	}
}
