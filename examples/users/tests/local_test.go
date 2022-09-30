package tests

import (
	"context"
	"testing"

	_ "github.com/contextcloud/eventstore/es/providers/data/pg"
	_ "github.com/contextcloud/eventstore/es/providers/stream/gpub"
	_ "github.com/contextcloud/eventstore/es/providers/stream/npub"

	"github.com/contextcloud/eventstore/es"
	"github.com/contextcloud/eventstore/examples/users/aggregates"
	"github.com/contextcloud/eventstore/examples/users/sagas"
	"go.opentelemetry.io/otel"
)

func Test_Local(t *testing.T) {
	shutdown, err := Zipkin()
	if err != nil {
		t.Error(err)
		return
	}

	pcfg, err := Provider()
	if err != nil {
		t.Error(err)
		return
	}

	cfg, err := es.NewConfig(
		pcfg,
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
	pcfg, err := Provider()
	if err != nil {
		b.Error(err)
		return
	}

	cfg, err := es.NewConfig(
		pcfg,
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
