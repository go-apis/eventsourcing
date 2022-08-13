package config

import (
	"github.com/contextcloud/eventstore/es"
	"github.com/contextcloud/eventstore/examples/users/aggregates"
	"github.com/contextcloud/eventstore/examples/users/sagas"
)

func SetupClient(conn es.Conn) (es.Client, error) {
	cfg, err := es.NewConfig(
		"users",
		"v0.1.0",
		&aggregates.User{},
		&aggregates.ExternalUser{},
		sagas.NewConnectionSaga(),
	)
	if err != nil {
		return nil, err
	}
	return es.NewClient(cfg, conn)
}
