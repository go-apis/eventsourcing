package tests

import (
	"context"
	"testing"

	"github.com/contextcloud/eventstore/es"
	"github.com/contextcloud/eventstore/examples/groups/data"
	"github.com/contextcloud/eventstore/examples/groups/data/commands"
	"github.com/contextcloud/eventstore/pkg/pgdb"
	"github.com/google/uuid"

	_ "github.com/contextcloud/eventstore/es/providers/data/pg"
	_ "github.com/contextcloud/eventstore/es/providers/stream/noop"
)

func TestHandler(t *testing.T) {
	pcfg := &es.ProviderConfig{
		Data: es.DataConfig{
			Type: "pg",
			Pg: &pgdb.Config{
				Host:     "localhost",
				Port:     5432,
				User:     "es",
				Password: "mysecret",
				Name:     "eventstore",
			},
		},
		Stream: es.StreamConfig{
			Type: "noop",
		},
	}

	ctx := context.Background()
	cli, err := data.NewClient(ctx, pcfg)
	if err != nil {
		t.Error(err)
		return
	}

	unit, err := cli.Unit(ctx)
	if err != nil {
		t.Error(err)
		return
	}
	tx, err := unit.NewTx(ctx)
	if err != nil {
		t.Error(err)
		return
	}

	uctx := es.SetUnit(ctx, unit)

	cli.HandleCommands(uctx, &commands.DemoCommand{
		BaseCommand: es.BaseCommand{
			AggregateId: uuid.New(),
		},
		Name: "demo",
	})

	mod, err := tx.Commit(uctx)
	if err != nil {
		t.Error(err)
		return
	}

	t.Log(mod)
}
