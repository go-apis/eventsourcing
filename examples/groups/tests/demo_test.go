package tests

import (
	"context"
	"testing"

	_ "github.com/contextcloud/eventstore/es/providers/data/pg"
	_ "github.com/contextcloud/eventstore/es/providers/stream/noop"
	"github.com/contextcloud/goutils/xgorm"

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
			Pg: &xgorm.DbConfig{
				Host:     "localhost",
				Port:     5432,
				Username: "es",
				Password: "es",
				Database: "es",
			},
			Reset: true,
		},
		Stream: es.StreamConfig{
			Type: "noop",
		},
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
