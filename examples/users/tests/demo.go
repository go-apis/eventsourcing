package tests

import (
	"context"
	"log"
	"os"

	"github.com/contextcloud/eventstore/es"
	"github.com/contextcloud/eventstore/es/filters"
	"github.com/contextcloud/eventstore/es/local"
	"github.com/contextcloud/eventstore/es/pb"
	"github.com/contextcloud/eventstore/examples/users/aggregates"
	"github.com/contextcloud/eventstore/examples/users/commands"
	"github.com/contextcloud/eventstore/pkg/db"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

var logger = log.New(os.Stderr, "zipkin-example", log.Ldate|log.Ltime|log.Llongfile)

type Shutdown = func(context.Context) error

func Zipkin() (Shutdown, error) {
	url := "http://localhost:9411/api/v2/spans"
	exporter, err := zipkin.New(
		url,
		zipkin.WithLogger(logger),
	)
	if err != nil {
		return nil, err
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("tests"),
		)),
	)
	otel.SetTracerProvider(tp)
	return tp.Shutdown, nil
}

func LocalConn() (es.Conn, error) {
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

func QueryUsers(ctx context.Context) error {
	userQuery := es.NewQuery[*aggregates.User]()
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

func UserCommands(ctx context.Context) error {
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

	unit, err := es.GetUnit(ctx)
	if err != nil {
		return err
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
