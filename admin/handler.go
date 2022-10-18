// go:generate go run github.com/99designs/gqlgen generate
package admin

import (
	"context"
	"fmt"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/contextcloud/eventstore/admin/graph"
	"github.com/contextcloud/eventstore/admin/graph/generated"
	"github.com/contextcloud/graceful/config"
	"github.com/go-chi/chi/v5"
)

func NewHandler(ctx context.Context, cfg *config.Config) (http.Handler, error) {
	schema := generated.NewExecutableSchema(generated.Config{
		Resolvers: &graph.Resolver{},
	})

	srv := handler.NewDefaultServer(schema)
	srv.SetRecoverFunc(func(ctx context.Context, err interface{}) error {
		return fmt.Errorf("graphql panic: %v", err)
	})
	srv.AddTransport(transport.POST{})
	srv.Use(extension.Introspection{})

	router := chi.NewRouter()
	router.Handle("/playground", playground.Handler("Eventstore", "/query"))
	router.Handle("/query", srv)

	return router, nil
}
