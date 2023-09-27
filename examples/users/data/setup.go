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
	esCfg, err := es.NewConfig(
		pcfg,
		&aggregates.StandardUser{},
		&aggregates.User{},
		&aggregates.ExternalUser{},
		sagas.NewConnectionSaga(),
		projectors.NewUserProjector(),
		&events.GroupAdded{},
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
