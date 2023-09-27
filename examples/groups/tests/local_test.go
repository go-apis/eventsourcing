package tests

import (
	"context"
	"testing"

	"github.com/contextcloud/eventstore/es"
	_ "github.com/contextcloud/eventstore/es/providers/data/pg"
	_ "github.com/contextcloud/eventstore/es/providers/stream/noop"
	"github.com/stretchr/testify/require"
)

func Test(t *testing.T) {
	tester, err := NewTester()
	require.NoError(t, err)

	t.Run("create", func(t *testing.T) {
		cli := tester.Client()

		ctx := context.Background()

		unit, err := cli.Unit(ctx)
		require.NoError(t, err)

		ctx = es.SetUnit(ctx, unit)
	})
}
