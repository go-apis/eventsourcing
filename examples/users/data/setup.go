package data

import (
	"context"

	"github.com/contextcloud/eventstore/es"
	"github.com/contextcloud/eventstore/examples/users/data/aggregates"
	"github.com/contextcloud/eventstore/examples/users/data/events"
	"github.com/contextcloud/eventstore/examples/users/data/projectors"
	"github.com/contextcloud/eventstore/examples/users/data/sagas"
)

func NewClient(ctx context.Context, pcfg *es.ProviderConfig) (es.Client, error) {
	if err := es.RegistryAdd(
		pcfg.Service,
		&aggregates.StandardUser{},
		&aggregates.User{},
		&aggregates.ExternalUser{},
		sagas.NewConnectionSaga(),
		projectors.NewUserProjector(),
		&events.GroupAdded{},
	); err != nil {
		return nil, err
	}

	cli, err := es.NewClient(ctx, pcfg)
	if err != nil {
		return nil, err
	}
	return cli, nil
}
