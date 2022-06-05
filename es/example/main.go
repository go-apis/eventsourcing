package main

import (
	"eventstore/es"
	"eventstore/es/example/aggregates"
	"eventstore/es/example/commands"
	"eventstore/es/httputil"
	"eventstore/es/local"
	"eventstore/es/local/pg"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	data, err := pg.NewData("postgresql://es:es@localhost:5432/eventstore?sslmode=disable")
	if err != nil {
		panic(err)
	}

	cli := local.NewClient(data, "example")
	store := es.NewEventStore(cli)
	store.AddCommandHandler(&aggregates.User{})

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})
	r.Post("/commands/demo", httputil.NewCommandHandler[commands.CreateUser](store))

	http.ListenAndServe(":3000", r)
}
