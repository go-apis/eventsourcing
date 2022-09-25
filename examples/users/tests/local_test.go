package tests

import (
	"context"
	"testing"

	"github.com/contextcloud/eventstore/es"
	"github.com/contextcloud/eventstore/es/providers"
	"github.com/contextcloud/eventstore/es/providers/data"
	"github.com/contextcloud/eventstore/es/providers/stream"
	"github.com/contextcloud/eventstore/examples/users/aggregates"
	"github.com/contextcloud/eventstore/examples/users/sagas"
	"github.com/contextcloud/eventstore/pkg/db"
	"github.com/contextcloud/eventstore/pkg/pub"
	"go.opentelemetry.io/otel"
)

func Test_Local(t *testing.T) {
	shutdown, err := Zipkin()
	if err != nil {
		t.Error(err)
	}
	if err := Reset(); err != nil {
		t.Error(err)
		return
	}

	pCfg := &providers.Config{
		ServiceName: "users",
		Version:     "v1",
		Data: data.Config{
			Type: "pg",
			Pg: &db.Config{
				Host:     "localhost",
				Port:     5432,
				User:     "es",
				Password: "es",
				Name:     "eventstore",
			},
		},
		Stream: stream.Config{
			Type: "gpub",
			Stream: &pub.Config{
				ProjectId: "nordic-gaming",
				TopicId:   "test_topic",
			},
		},
	}

	cfg, err := es.NewConfig(
		pCfg,
		&aggregates.User{},
		&aggregates.ExternalUser{},
		sagas.NewConnectionSaga(),
	)
	if err != nil {
		t.Error(err)
		return
	}

	cli, err := es.NewClient(cfg)
	if err != nil {
		t.Error(err)
		return
	}

	ctx := context.Background()
	pctx, pspan := otel.Tracer("test").Start(ctx, "Local")
	defer pspan.End()

	if err := cli.Initialize(pctx); err != nil {
		t.Error(err)
		return
	}

	// the event store should know the aggregates and the commands.
	unit, err := cli.Unit(pctx)
	if err != nil {
		t.Error(err)
		return
	}
	pctx = es.SetUnit(pctx, unit)

	userId, _, err := UserCommands(pctx)
	if err != nil {
		t.Error(err)
		return
	}

	for i := 0; i < 1000; i++ {
		if err := QueryUsers(pctx, userId); err != nil {
			t.Error(err)
			return
		}

		t.Logf("index: %d", i)
	}

	shutdown(pctx)
}

func Benchmark_CreateUsers(b *testing.B) {
	if err := Reset(); err != nil {
		b.Error(err)
		return
	}

	pCfg := &providers.Config{
		ServiceName: "users",
		Version:     "v1",
		Data: data.Config{
			Type: "pg",
			Pg: &db.Config{
				Host:     "localhost",
				Port:     5432,
				User:     "es",
				Password: "es",
				Name:     "eventstore",
			},
		},
		Stream: stream.Config{
			Type: "gpub",
			Stream: &pub.Config{
				ProjectId: "nordic-gaming",
				TopicId:   "test_topic",
			},
		},
	}

	cfg, err := es.NewConfig(
		pCfg,
		&aggregates.User{},
		&aggregates.ExternalUser{},
		sagas.NewConnectionSaga(),
	)
	if err != nil {
		b.Error(err)
		return
	}

	cli, err := es.NewClient(cfg)
	if err != nil {
		b.Error(err)
		return
	}

	pctx := context.Background()
	if err := cli.Initialize(pctx); err != nil {
		b.Error(err)
		return
	}

	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		// the event store should know the aggregates and the commands.
		unit, err := cli.Unit(ctx)
		if err != nil {
			b.Error(err)
			return
		}
		ctx = es.SetUnit(ctx, unit)
		if _, _, err := UserCommands(ctx); err != nil {
			b.Error(err)
			return
		}
	}
}
