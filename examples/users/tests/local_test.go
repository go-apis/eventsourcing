package tests

import (
	"context"
	"testing"

	"github.com/contextcloud/eventstore/es"
	"github.com/contextcloud/eventstore/examples/users/config"
	"go.opentelemetry.io/otel"
)

func Test_Local(t *testing.T) {
	shutdown, err := Zipkin()
	if err != nil {
		t.Error(err)
	}

	conn, err := LocalConn()
	if err != nil {
		t.Error(err)
		return
	}

	cli, err := config.SetupClient(conn)
	if err != nil {
		t.Error(err)
		return
	}

	ctx := context.Background()
	pctx, pspan := otel.Tracer("test").Start(ctx, "Local")
	defer pspan.End()

	// the event store should know the aggregates and the commands.
	unit, err := cli.Unit(pctx)
	if err != nil {
		t.Error(err)
		return
	}
	pctx = es.SetUnit(pctx, unit)

	if err := UserCommands(pctx); err != nil {
		t.Error(err)
		return
	}

	for i := 0; i < 1000; i++ {
		if err := QueryUsers(pctx); err != nil {
			t.Error(err)
			return
		}

		t.Logf("index: %d", i)
	}

	shutdown(pctx)
}
