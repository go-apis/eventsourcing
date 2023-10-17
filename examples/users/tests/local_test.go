package tests

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/contextcloud/eventstore/es"
	"github.com/contextcloud/eventstore/es/filters"
	"github.com/contextcloud/eventstore/examples/users/data/aggregates"
	"github.com/contextcloud/eventstore/examples/users/data/commands"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func Test(t *testing.T) {
	tester, err := NewTester()
	require.NoError(t, err)

	t.Run("create", func(t *testing.T) {
		cli := tester.Client()

		ctx := context.Background()
		unit, errU := cli.Unit(ctx)
		require.NoError(t, errU)

		ctx = es.SetUnit(ctx, unit)

		userId1 := uuid.New()
		userId2 := uuid.New()
		cmds := []es.Command{
			&commands.CreateUser{
				BaseCommand: es.BaseCommand{
					AggregateId: userId1,
				},
				Username: "chris.kolenko",
				Password: "12345678",
			},
			&commands.AddEmail{
				BaseCommand: es.BaseCommand{
					AggregateId: userId1,
				},
				Email: "chris@context.gg",
			},
			&commands.AddConnection{
				BaseCommand: es.BaseCommand{
					AggregateId: userId1,
				},
				Name:     "Smashgg",
				UserId:   "demo1",
				Username: "chris.kolenko",
			},
			&commands.UpdateConnection{
				BaseCommand: es.BaseCommand{
					AggregateId: userId1,
				},
				Username: "aaaaaaaaaa",
			},
			&commands.CreateUser{
				BaseCommand: es.BaseCommand{
					AggregateId: userId2,
				},
				BaseNamespaceCommand: es.BaseNamespaceCommand{
					Namespace: "other",
				},
				Username: "calvin.harris",
				Password: "12345678",
			},
			&commands.DeleteUser{
				BaseCommand: es.BaseCommand{
					AggregateId: userId2,
				},
			},
			&commands.AddGroup{
				BaseCommand: es.BaseCommand{
					AggregateId: userId2,
				},
				GroupId: uuid.New(),
			},
		}

		errD := unit.Dispatch(ctx, cmds...)
		require.NoError(t, errD)

		userQuery := es.NewQuery[*aggregates.User]()
		user, err := userQuery.Get(ctx, userId1)
		require.NoError(t, err)

		filter := filters.Filter{
			Where: filters.WhereClause{
				Column: "username",
				Op:     "eq",
				Args:   []interface{}{"chris.kolenko"},
			},
			Order:  []filters.Order{{Column: "username"}},
			Limit:  filters.Limit(1),
			Offset: filters.Offset(0),
		}

		users, err := userQuery.Find(ctx, filter)
		require.NoError(t, err)

		total, err := userQuery.Count(ctx, filter)
		require.NoError(t, err)

		log.Printf("user: %+v", user)
		log.Printf("users: %+v", users)
		log.Printf("total: %+v", total)

		time.Sleep(5 * time.Second)
	})
}
