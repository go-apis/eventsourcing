package config

import (
	"github.com/contextcloud/eventstore/es"
	"github.com/contextcloud/eventstore/examples/groups/aggregates"
	"github.com/contextcloud/eventstore/examples/groups/sagas"
)

func EventStoreConfig() (es.Config, error) {
	cfg, err := es.NewConfig(
		"groups",
		"v0.1.0",
		&aggregates.Group{},
		sagas.NewUserSaga(),
	)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
