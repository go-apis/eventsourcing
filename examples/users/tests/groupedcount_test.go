package tests

import (
	"context"
	"testing"

	"github.com/go-apis/eventsourcing/es"
	"github.com/go-apis/eventsourcing/examples/users/data/aggregates"
	"github.com/go-apis/eventsourcing/examples/users/data/commands"
	"github.com/go-apis/eventsourcing/examples/users/helpers"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// TestGroupedCount exercises the real grouped aggregation against the sqlite
// provider: grouping correctness, the count per group, and the injection guard.
func TestGroupedCount(t *testing.T) {
	tester, err := NewTester()
	require.NoError(t, err)

	cli := tester.Client()
	ctx := context.Background()
	unit, err := cli.Unit(ctx)
	require.NoError(t, err)

	ctx = es.SetUnit(ctx, unit)
	ctx = helpers.SetSkipSaga(ctx)

	mk := func(username string) es.Command {
		return &commands.CreateUser{
			BaseCommand: es.BaseCommand{AggregateId: uuid.New()},
			Username:    username,
			Password:    "12345678",
		}
	}
	require.NoError(t, unit.Dispatch(ctx,
		mk("alpha"), mk("alpha"), mk("alpha"), mk("beta"), mk("beta"),
	))

	userQuery := es.NewQuery[*aggregates.User]()

	groups, err := userQuery.GroupedCount(ctx, es.Filter{}, "username")
	require.NoError(t, err)

	counts := map[string]int{}
	total := 0
	for _, g := range groups {
		counts[g.Key] = g.Count
		total += g.Count
	}
	require.Equal(t, 3, counts["alpha"])
	require.Equal(t, 2, counts["beta"])
	require.Equal(t, 5, total)

	// a where filter narrows the grouped set
	filtered, err := userQuery.GroupedCount(ctx, es.Filter{
		Where: es.WhereClause{Column: "username", Op: "eq", Args: []interface{}{"alpha"}},
	}, "username")
	require.NoError(t, err)
	require.Len(t, filtered, 1)
	require.Equal(t, "alpha", filtered[0].Key)
	require.Equal(t, 3, filtered[0].Count)

	// injection guard: a non-identifier group-by column is rejected
	_, err = userQuery.GroupedCount(ctx, es.Filter{}, "username); drop table users;--")
	require.Error(t, err)

	// Distinct is explicitly unsupported
	_, err = userQuery.GroupedCount(ctx, es.Filter{Distinct: []interface{}{"username"}}, "username")
	require.Error(t, err)
}
