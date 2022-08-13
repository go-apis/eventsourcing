package tests

import (
	"context"
	"testing"

	"github.com/contextcloud/eventstore/es"
	"github.com/contextcloud/eventstore/examples/users/config"
)

func Test_Pb(t *testing.T) {
	shutdown, err := Zipkin()
	if err != nil {
		t.Error(err)
	}

	conn, err := PbConn()
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

	// the event store should know the aggregates and the commands.
	unit, err := cli.Unit(ctx)
	if err != nil {
		t.Error(err)
		return
	}
	ctx = es.SetUnit(ctx, unit)

	if err := UserCommands(ctx); err != nil {
		t.Error(err)
		return
	}

	for i := 0; i < 1000; i++ {
		if err := QueryUsers(ctx); err != nil {
			t.Error(err)
			return
		}

		t.Logf("index: %d", i)
	}

	shutdown(ctx)
}
