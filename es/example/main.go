package main

import (
	"eventstore/es"
	"eventstore/es/example/commands"
	"eventstore/es/httputil"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	store := es.NewEventStore(nil)

	r := chi.NewRouter()
	r.Use(httputil.Transaction(store))
	r.Use(middleware.Logger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})
	r.Post("/commands/demo", httputil.NewCommandHandler[commands.CreateUser](store))

	http.ListenAndServe(":3000", r)
}
