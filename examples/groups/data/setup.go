package data

import (
	"context"

	"github.com/go-apis/eventsourcing/es"
	"github.com/go-apis/eventsourcing/examples/groups/data/aggregates"
	"github.com/go-apis/eventsourcing/examples/groups/data/commands"
	"github.com/go-apis/eventsourcing/examples/groups/data/handlers"
	"github.com/go-apis/eventsourcing/examples/groups/data/sagas"
)

func NewClient(ctx context.Context, pcfg *es.ProviderConfig) (es.Client, error) {
	reg, err := es.NewRegistry(
		pcfg.Service,
		&aggregates.Group{},
		&aggregates.Demo{},
		handlers.NewDemoHandler(),
		sagas.NewDemoSaga(),
		sagas.NewUserSaga(),
		es.NewAggregateConfig(
			&aggregates.Community{},
			es.EntitySnapshotEvery(1),
			&commands.CommunityNewCommand{},
			&commands.CommunityDeleteCommand{},
		),
	)
	if err != nil {
		return nil, err
	}

	cli, err := es.NewClient(ctx, pcfg, reg)
	if err != nil {
		return nil, err
	}
	return cli, nil
}
