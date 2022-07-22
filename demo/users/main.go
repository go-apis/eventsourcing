package main

import (
	"encoding/json"
	"net/http"

	"github.com/contextcloud/eventstore/demo/users/aggregates"
	"github.com/contextcloud/eventstore/demo/users/commands"
	"github.com/contextcloud/eventstore/demo/users/sagas"
	"github.com/contextcloud/eventstore/es/filters"
	"github.com/contextcloud/eventstore/es/local"
	"github.com/contextcloud/eventstore/pkg/db"

	"github.com/contextcloud/eventstore/es"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func handle[T es.Command](cli es.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		unit, err := cli.NewUnit(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		commander := es.NewCommander[T](unit)
		commander.Handle(w, r)
	}
}

func userQueryFunc(cli es.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		unit, err := cli.NewUnit(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		filter := filters.Filter{}

		q := es.NewQuery[aggregates.User](unit)
		out, err := q.Find(ctx, filter)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := json.NewEncoder(w).Encode(out); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
}

func main() {
	conn, err := local.NewConn(
		db.WithDbUser("es"),
		db.WithDbPassword("es"),
		db.WithDbName("eventstore"),
	)
	if err != nil {
		panic(err)
	}

	cfg, err := es.NewConfig(
		"example",
		&aggregates.User{},
		&aggregates.ExternalUser{},
		sagas.NewConnectionSaga(),
	)
	if err != nil {
		panic(err)
	}
	cli, err := es.NewClient(cfg, conn)
	if err != nil {
		panic(err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})
	r.Post("/commands/demo", handle[commands.CreateUser](cli))
	r.Get("/users", userQueryFunc(cli))

	http.ListenAndServe(":3000", r)
}
