package tests

import (
	"context"
	"log"
	"os"

	"github.com/contextcloud/eventstore/es"
	"github.com/contextcloud/eventstore/es/filters"
	"github.com/contextcloud/eventstore/examples/users/aggregates"
	"github.com/contextcloud/eventstore/examples/users/commands"
	"github.com/contextcloud/eventstore/pkg/gcppubsub"
	"github.com/contextcloud/eventstore/pkg/natspubsub"
	"github.com/contextcloud/eventstore/pkg/pgdb"
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

func Provider() (*es.ProviderConfig, error) {
	pcfg := &es.ProviderConfig{
		ServiceName: "users",
		Version:     "v1",
		Data: es.DataConfig{
			Type: "pg",
			Pg: &pgdb.Config{
				Debug:    true,
				Host:     "localhost",
				Port:     5432,
				User:     "es",
				Password: "es",
				Name:     "eventstore",
			},
		},
		Stream: es.StreamConfig{
			Type: "noop",
			PubSub: &gcppubsub.Config{
				ProjectId: "nordic-gaming",
				TopicId:   "test_topic",
			},
			Nats: &natspubsub.Config{
				Url:     "nats://localhost:4222",
				Subject: "test",
			},
		},
	}
	if err := pgdb.Reset(pcfg.Data.Pg); err != nil {
		return nil, err
	}
	// if err := gcppubsub.Reset(pcfg.Stream.PubSub); err != nil {
	// 	return nil, err
	// }
	return pcfg, nil
}

func QueryUsers(ctx context.Context, userId uuid.UUID) error {
	userQuery := es.NewQuery[*aggregates.User]()
	user, err := userQuery.Get(ctx, userId)
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

func UserCommands(ctx context.Context) (uuid.UUID, uuid.UUID, error) {
	userId1 := uuid.New()
	userId2 := uuid.New()

	cmds := []es.Command{
		&commands.CreateUser{
			BaseCommand: es.BaseCommand{
				AggregateId: userId1,
			},
			Username: "chris.kolenko",
			Password: "12345678",
		},
		&commands.AddEmail{
			BaseCommand: es.BaseCommand{
				AggregateId: userId1,
			},
			Email: "chris@context.gg",
		},
		&commands.AddConnection{
			BaseCommand: es.BaseCommand{
				AggregateId: userId1,
			},
			Name:     "Smashgg",
			UserId:   "demo1",
			Username: "chris.kolenko",
		},
		&commands.UpdateConnection{
			BaseCommand: es.BaseCommand{
				AggregateId: userId1,
			},
			Username: "aaaaaaaaaa",
		},
		&commands.CreateUser{
			BaseCommand: es.BaseCommand{
				AggregateId: userId2,
			},
			BaseNamespaceCommand: es.BaseNamespaceCommand{
				Namespace: "other",
			},
			Username: "calvin.harris",
			Password: "12345678",
		},
	}

	unit, err := es.GetUnit(ctx)
	if err != nil {
		return userId1, userId2, err
	}

	tx, err := unit.NewTx(ctx)
	if err != nil {
		return userId1, userId2, err
	}
	defer tx.Rollback(ctx)

	// send a commands to store.
	if err := unit.Dispatch(ctx, cmds...); err != nil {
		return userId1, userId2, err
	}

	// commit the tx.
	if _, err := tx.Commit(ctx); err != nil {
		return userId1, userId2, err
	}

	return userId1, userId2, nil
}
