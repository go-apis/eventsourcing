package users

import (
	"context"
	"encoding/json"
	"net/http"

	_ "github.com/contextcloud/eventstore/es/providers/data/pg"
	_ "github.com/contextcloud/eventstore/es/providers/stream/gpub"
	_ "github.com/contextcloud/eventstore/es/providers/stream/npub"

	"github.com/contextcloud/eventstore/es/filters"
	"github.com/contextcloud/eventstore/examples/users/aggregates"
	"github.com/contextcloud/eventstore/examples/users/commands"
	"github.com/contextcloud/eventstore/examples/users/sagas"
	"github.com/contextcloud/graceful/config"
	"github.com/riandyrn/otelchi"

	"github.com/contextcloud/eventstore/es"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

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
	pCfg := &es.ProviderConfig{}
	if err := cfg.Parse(pCfg); err != nil {
		return nil, err
	}

	esCfg, err := es.NewConfig(
		pCfg,
		&aggregates.User{},
		&aggregates.ExternalUser{},
		sagas.NewConnectionSaga(),
	)
	if err != nil {
		return nil, err
	}

	cli, err := es.NewClient(ctx, esCfg)
	if err != nil {
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
