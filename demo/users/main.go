package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/contextcloud/eventstore/demo/users/aggregates"
	"github.com/contextcloud/eventstore/demo/users/commands"
	"github.com/contextcloud/eventstore/demo/users/sagas"
	"github.com/contextcloud/eventstore/es/filters"
	"github.com/contextcloud/eventstore/es/local"
	"github.com/contextcloud/eventstore/pkg/db"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"

	"github.com/contextcloud/eventstore/es"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func CreateUnit(cli es.Client) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			unit, err := cli.Unit(ctx)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			ctx = es.SetUnit(ctx, unit)
			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func handle[T es.Command](cli es.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		unit, err := es.GetUnit(ctx)
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

		filter := filters.Filter{}

		q := es.NewQuery[*aggregates.User]()
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

var logger = log.New(os.Stderr, "zipkin-example", log.Ldate|log.Ltime|log.Llongfile)

func main() {
	url := "http://localhost:9411/api/v2/spans"
	exporter, err := zipkin.New(
		url,
		zipkin.WithLogger(logger),
	)
	if err != nil {
		panic(err)
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("zipkin-test"),
		)),
	)
	otel.SetTracerProvider(tp)
	defer tp.Shutdown(context.Background())

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
		"v1.0.0",
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
	r.Use(CreateUnit(cli))
	r.Use(middleware.Logger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})
	r.Post("/commands/demo", handle[commands.CreateUser](cli))
	r.Get("/users", userQueryFunc(cli))

	http.ListenAndServe(":3000", r)
}
