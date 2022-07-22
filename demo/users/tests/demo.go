package tests

import (
	"context"
	"log"

	"github.com/contextcloud/eventstore/demo/users/aggregates"
	"github.com/contextcloud/eventstore/demo/users/commands"
	"github.com/contextcloud/eventstore/es"
	"github.com/contextcloud/eventstore/es/filters"
	"github.com/contextcloud/eventstore/es/local"
	"github.com/contextcloud/eventstore/es/pb"
	"github.com/contextcloud/eventstore/pkg/db"
	"github.com/google/uuid"
)

func LocalConn() (es.Conn, error) {
	// dsn := "postgresql://inflow:kU1tvu@pg.data:5432/inflow-assets?sslmode=disable"
	// dsn := "postgresql://es:es@localhost:5432/eventstore?sslmode=disable"

	opts := []db.OptionFunc{
		db.WithDbUser("es"),
		db.WithDbPassword("es"),
		db.WithDbName("eventstore"),
	}

	if err := db.Reset(opts...); err != nil {
		return nil, err
	}

	return local.NewConn(opts...)
}

func PbConn() (es.Conn, error) {
	dsn := "localhost:3332"
	return pb.NewConn(dsn)
}

func QueryUsers(ctx context.Context, unit es.Unit) error {
	userQuery := es.NewQuery[aggregates.User](unit)
	user, err := userQuery.Load(ctx, uuid.MustParse("98f1f7d3-f312-4d57-8847-5b9ac8d5797d"))
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

func UserCommands(ctx context.Context, unit es.Unit) error {
	cmds := []es.Command{
		&commands.CreateUser{
			BaseCommand: es.BaseCommand{
				AggregateId: uuid.MustParse("98f1f7d3-f312-4d57-8847-5b9ac8d5797d"),
			},
			Username: "chris.kolenko",
			Password: "12345678",
		},
		&commands.AddEmail{
			BaseCommand: es.BaseCommand{
				AggregateId: uuid.MustParse("98f1f7d3-f312-4d57-8847-5b9ac8d5797d"),
			},
			Email: "chris@context.gg",
		},
		&commands.AddConnection{
			BaseCommand: es.BaseCommand{
				AggregateId: uuid.MustParse("98f1f7d3-f312-4d57-8847-5b9ac8d5797d"),
			},
			Name:     "Smashgg",
			UserId:   "demo1",
			Username: "chris.kolenko",
		},
		&commands.UpdateConnection{
			BaseCommand: es.BaseCommand{
				AggregateId: uuid.MustParse("98f1f7d3-f312-4d57-8847-5b9ac8d5797d"),
			},
			Username: "aaaaaaaaaa",
		},
		&commands.CreateUser{
			BaseCommand: es.BaseCommand{
				AggregateId: uuid.MustParse("2ca16492-ea7a-4d96-8599-b256c26e89b5"),
			},
			Username: "calvin.harris",
			Password: "12345678",
		},
	}

	tx, err := unit.NewTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// send a commands to store.
	if err := unit.Dispatch(ctx, cmds...); err != nil {
		return err
	}

	// commit the tx.
	if _, err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}
