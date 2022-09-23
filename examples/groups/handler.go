package groups

import (
	"context"
	"net/http"

	"github.com/contextcloud/eventstore/es"
	"github.com/contextcloud/eventstore/es/gstream"
	"github.com/contextcloud/eventstore/es/local"
	"github.com/contextcloud/eventstore/examples/groups/aggregates"
	"github.com/contextcloud/eventstore/examples/groups/commands"
	"github.com/contextcloud/eventstore/examples/groups/events"
	"github.com/contextcloud/eventstore/examples/groups/sagas"
	"github.com/contextcloud/eventstore/pkg/db"
	"github.com/contextcloud/eventstore/pkg/pub/g"
	"github.com/contextcloud/graceful/config"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/riandyrn/otelchi"
)

type Config struct {
	Db struct {
		Host     string
		Port     int
		User     string
		Password string
		Name     string
	}
	Streamer struct {
		Project string
		Topic   string
	}
}

func NewHandler(ctx context.Context, cfg *config.Config) (http.Handler, error) {
	ourCfg := &Config{}
	if err := cfg.Parse(ourCfg); err != nil {
		return nil, err
	}

	esCfg, err := es.NewConfig(
		cfg.ServiceName,
		cfg.Version,
		&aggregates.Group{},
		sagas.NewUserSaga(),
		es.NewAggregateConfig(
			&aggregates.Community{},
			es.EntityDisableProject(),
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

	conn, err := local.NewConn(
		db.WithDbHost(ourCfg.Db.Host),
		db.WithDbPort(ourCfg.Db.Port),
		db.WithDbUser(ourCfg.Db.User),
		db.WithDbPassword(ourCfg.Db.Password),
		db.WithDbName(ourCfg.Db.Name),
	)
	if err != nil {
		return nil, err
	}

	cli, err := g.Open(
		g.WithProjectId(ourCfg.Streamer.Project),
		g.WithTopicId(ourCfg.Streamer.Topic),
	)

	streamer, err := gstream.NewStreamer(cli)
	if err != nil {
		return nil, err
	}

	cli, err := es.NewClient(esCfg, conn, streamer)
	if err != nil {
		return nil, err
	}

	if err := cli.Initialize(ctx); err != nil {
		return nil, err
	}

	r := chi.NewRouter()
	r.Use(otelchi.Middleware("server", otelchi.WithChiRoutes(r)))
	r.Use(es.CreateUnit(cli))
	r.Use(middleware.Logger)
	r.Post("/commands/newcommunity", es.NewCommander[*commands.CommunityNewCommand]())
	r.Post("/commands/creategroup", es.NewCommander[*commands.CreateGroup]())

	return r, nil
}
