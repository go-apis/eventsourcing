package tests

import (
	"context"
	"log"
	"testing"

	"github.com/contextcloud/eventstore/es/example/aggregates"
	"github.com/contextcloud/eventstore/es/example/commands"
	"github.com/contextcloud/eventstore/es/example/sagas"
	"github.com/contextcloud/eventstore/es/filters"
	"github.com/contextcloud/eventstore/es/local/g"

	"github.com/contextcloud/eventstore/es"
)

func Setup() (es.Client, error) {
	dsn := "postgresql://es:es@localhost:5432/eventstore?sslmode=disable"
	if err := g.ResetDb(dsn); err != nil {
		return nil, err
	}

	data, err := g.NewData(dsn)
	if err != nil {
		return nil, err
	}

	cfg, err := es.NewConfig(
		"example",
		&aggregates.User{},
		&aggregates.ExternalUser{},
		sagas.NewConnectionSaga(),
	)
	if err != nil {
		return nil, err
	}

	if err := data.Initialize(cfg); err != nil {
		return nil, err
	}

	cli := es.NewClient(cfg, data)
	return cli, nil
}

func QueryUsers(cli es.Client) error {
	ctx := context.Background()

	// the event store should know the aggregates and the commands.
	unit, err := cli.NewUnit(ctx)
	if err != nil {
		return err
	}
	defer unit.Rollback(ctx)

	userQuery := es.NewQuery[aggregates.User](unit)
	user, err := userQuery.Load(ctx, "98f1f7d3-f312-4d57-8847-5b9ac8d5797d")
	if err != nil {
		return err
	}

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
		return err
	}

	total, err := userQuery.Count(ctx, filter)
	if err != nil {
		return err
	}

	log.Printf("user: %+v", user)
	log.Printf("users: %+v", users)
	log.Printf("total: %+v", total)

	return err
}

func Test_It(t *testing.T) {
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

	cli, err := Setup()
	if err != nil {
		t.Error(err)
		return
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

	// commit the tx.
	if err := unit.Commit(ctx); err != nil {
		t.Error(err)
		return
	}

	for i := 0; i < 1000; i++ {
		if err := QueryUsers(cli); err != nil {
			t.Error(err)
			return
		}

		t.Logf("index: %d", i)
	}
}
