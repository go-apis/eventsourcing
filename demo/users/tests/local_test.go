package tests

import (
	"context"
	"testing"

	"github.com/contextcloud/eventstore/demo/users/config"
)

func Test_Local(t *testing.T) {
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

	// the event store should know the aggregates and the commands.
	unit, err := cli.NewUnit(ctx)
	if err != nil {
		t.Error(err)
		return
	}

	if err := UserCommands(ctx, unit); err != nil {
		t.Error(err)
		return
	}

	for i := 0; i < 1000; i++ {
		if err := QueryUsers(ctx, unit); err != nil {
			t.Error(err)
			return
		}

		t.Logf("index: %d", i)
	}
}
