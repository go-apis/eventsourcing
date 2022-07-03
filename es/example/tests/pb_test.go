package tests

import (
	"context"
	"testing"
)

func Test_Pb(t *testing.T) {
	conn, err := PbConn()
	if err != nil {
		t.Error(err)
		return
	}

	cli, err := SetupClient(conn)
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
