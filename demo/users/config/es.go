package config

import (
	"github.com/contextcloud/eventstore/demo/users/aggregates"
	"github.com/contextcloud/eventstore/demo/users/sagas"
	"github.com/contextcloud/eventstore/es"
)

func SetupClient(conn es.Conn) (es.Client, error) {
	cfg, err := es.NewConfig(
		"example",
		&aggregates.User{},
		&aggregates.ExternalUser{},
		sagas.NewConnectionSaga(),
	)
	if err != nil {
		return nil, err
	}
	return es.NewClient(cfg, conn)
}
