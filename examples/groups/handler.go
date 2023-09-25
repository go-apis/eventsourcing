package groups

import (
	"context"
	"net/http"

	_ "github.com/contextcloud/eventstore/es/providers/data/pg"
	_ "github.com/contextcloud/eventstore/es/providers/stream/gpub"
	_ "github.com/contextcloud/eventstore/es/providers/stream/npub"

	"github.com/contextcloud/eventstore/es"
	"github.com/contextcloud/eventstore/examples/groups/data"
	"github.com/contextcloud/graceful/config"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/riandyrn/otelchi"
)

func NewHandler(ctx context.Context, cfg *config.Config) (http.Handler, error) {
	pCfg := &es.ProviderConfig{}
	if err := cfg.Parse(pCfg); err != nil {
		return nil, err
	}

	cli, err := data.NewClient(ctx, pCfg)
	if err != nil {
		return nil, err
	}

	r := chi.NewRouter()
	r.Use(otelchi.Middleware("server", otelchi.WithChiRoutes(r)))
	r.Use(es.CreateUnit(cli))
	r.Use(middleware.Logger)

	return r, nil
}
