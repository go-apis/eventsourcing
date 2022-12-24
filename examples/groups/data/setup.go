package data

import (
	"context"

	"github.com/contextcloud/eventstore/es"
	"github.com/contextcloud/eventstore/examples/groups/data/aggregates"
	"github.com/contextcloud/eventstore/examples/groups/data/commands"
	"github.com/contextcloud/eventstore/examples/groups/data/events"
	"github.com/contextcloud/eventstore/examples/groups/data/handlers"
	"github.com/contextcloud/eventstore/examples/groups/data/sagas"
)

func NewClient(ctx context.Context, pcfg *es.ProviderConfig) (es.Client, error) {
	esCfg, err := es.NewConfig(
		pcfg,
		&aggregates.Group{},
		&aggregates.Demo{},
		handlers.NewDemoHandler(),
		sagas.NewDemoSaga(),
		sagas.NewUserSaga(),
		es.NewAggregateConfig(
			&aggregates.Community{},
			es.EntitySnapshotEvery(1),
			es.EntityEventTypes(
				&events.CommunityCreated{},
				&events.CommunityDeleted{},
				&events.CommunityStaffAdded{},
			),
			&commands.CommunityNewCommand{},
			&commands.CommunityDeleteCommand{},
		),
	)
	if err != nil {
		return nil, err
	}

	cli, err := es.NewClient(ctx, esCfg)
	if err != nil {
		return nil, err
	}
	return cli, nil
}
