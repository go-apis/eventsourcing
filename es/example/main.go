package main

import (
	"net/http"

	"github.com/contextcloud/eventstore/es/example/aggregates"
	"github.com/contextcloud/eventstore/es/example/commands"
	"github.com/contextcloud/eventstore/es/example/sagas"
	"github.com/contextcloud/eventstore/es/httputil"
	"github.com/contextcloud/eventstore/es/local/g"

	"github.com/contextcloud/eventstore/es"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	cfg, err := es.NewConfig(
		"example",
		&aggregates.User{},
		&aggregates.ExternalUser{},
		sagas.NewConnectionSaga(),
	)

	data, err := g.NewData(cfg, "postgresql://es:es@localhost:5432/eventstore?sslmode=disable")
	if err != nil {
		panic(err)
	}

	cli := es.NewClient(cfg, data)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})
	r.Post("/commands/demo", httputil.NewCommandHandler[commands.CreateUser](cli))
	r.Get("/users", httputil.NewQueryHandler[aggregates.User](cli))

	http.ListenAndServe(":3000", r)
}
