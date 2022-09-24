package users

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/contextcloud/eventstore/es/filters"
	"github.com/contextcloud/eventstore/es/gstream"
	"github.com/contextcloud/eventstore/es/local"
	"github.com/contextcloud/eventstore/examples/users/aggregates"
	"github.com/contextcloud/eventstore/examples/users/commands"
	"github.com/contextcloud/eventstore/examples/users/sagas"
	"github.com/contextcloud/eventstore/pkg/db"
	"github.com/contextcloud/eventstore/pkg/pub"
	"github.com/contextcloud/graceful/config"
	"github.com/riandyrn/otelchi"

	"github.com/contextcloud/eventstore/es"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Config struct {
	Db       db.Config
	Streamer gstream.Config
}

func userQueryFunc(cli es.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		filter := filters.Filter{}

		q := es.NewQuery[*aggregates.User]()
		out, err := q.Find(ctx, filter)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(out); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
}

func NewHandler(ctx context.Context, cfg *config.Config) (http.Handler, error) {
	ourCfg := &Config{}
	if err := cfg.Parse(ourCfg); err != nil {
		return nil, err
	}

	gormDb, err := db.Open(&ourCfg.Db)
	if err != nil {
		return nil, err
	}

	conn, err := local.NewConn(gormDb)
	if err != nil {
		return nil, err
	}

	gpub, err := gstream.Open(&ourCfg.Streamer)
	if err != nil {
		return nil, err
	}

	streamer, err := pub.NewStreamer(gpub)
	if err != nil {
		return nil, err
	}

	esCfg, err := es.NewConfig(
		cfg.ServiceName,
		cfg.Version,
		&aggregates.User{},
		&aggregates.ExternalUser{},
		sagas.NewConnectionSaga(),
	)
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
	r.Use(otelchi.Middleware(cfg.ServiceName, otelchi.WithChiRoutes(r)))
	r.Use(es.CreateUnit(cli))
	r.Use(middleware.Logger)
	r.Post("/commands/createuser", es.NewCommander[*commands.CreateUser]())
	r.Post("/commands/addgroup", es.NewCommander[*commands.AddGroup]())
	r.Get("/users", userQueryFunc(cli))

	return r, nil
}
