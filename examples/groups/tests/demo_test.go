package tests

import (
	"context"
	"testing"

	_ "github.com/contextcloud/eventstore/es/providers/data/pg"
	_ "github.com/contextcloud/eventstore/es/providers/stream/noop"
	"github.com/contextcloud/eventstore/pkg/pgdb"

	"github.com/contextcloud/eventstore/es"
	"github.com/contextcloud/eventstore/examples/groups/data"
	"github.com/contextcloud/eventstore/examples/groups/data/commands"
	"github.com/google/uuid"
)

func TestIt(t *testing.T) {
	ctx := context.Background()
	pcfg := &es.ProviderConfig{
		ServiceName: "groups",
		Version:     "v1",
		Data: es.DataConfig{
			Type: "pg",
			Pg: &pgdb.Config{
				Debug:    true,
				Host:     "localhost",
				Port:     5432,
				User:     "es",
				Password: "es",
				Name:     "es",
			},
		},
		Stream: es.StreamConfig{
			Type: "noop",
		},
	}

	if err := pgdb.Reset(pcfg.Data.Pg); err != nil {
		t.Fatal(err)
	}

	cli, err := data.NewClient(ctx, pcfg)
	if err != nil {
		t.Fatal(err)
	}

	unit, err := cli.Unit(ctx)
	if err != nil {
		t.Fatal(err)
	}
	tx, err := unit.NewTx(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Commit(ctx)

	if err := unit.Dispatch(ctx, &commands.NewDemo{
		BaseCommand: es.BaseCommand{
			AggregateId: uuid.New(),
		},
		Name: "demo",
	}); err != nil {
		t.Fatal(err)
	}
}
