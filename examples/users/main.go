package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/contextcloud/eventstore/es/filters"
	"github.com/contextcloud/eventstore/es/local"
	"github.com/contextcloud/eventstore/examples/users/aggregates"
	"github.com/contextcloud/eventstore/examples/users/commands"
	"github.com/contextcloud/eventstore/examples/users/config"
	"github.com/contextcloud/eventstore/pkg/db"
	"github.com/riandyrn/otelchi"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"

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

		if err := json.NewEncoder(w).Encode(out); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
}

func setupTracer() func(context.Context) error {
	logger := log.New(os.Stderr, "zipkin-example", log.Ldate|log.Ltime|log.Llongfile)
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
			semconv.ServiceNameKey.String("users"),
		)),
	)
	otel.SetTracerProvider(tp)
	return tp.Shutdown
}

func main() {
	shutdown := setupTracer()
	defer shutdown(context.Background())

	conn, err := local.NewConn(
		db.WithDbUser("es"),
		db.WithDbPassword("es"),
		db.WithDbName("eventstore"),
	)
	if err != nil {
		panic(err)
	}

	cli, err := config.SetupClient(conn)
	if err != nil {
		panic(err)
	}

	r := chi.NewRouter()
	r.Use(otelchi.Middleware("server", otelchi.WithChiRoutes(r)))
	r.Use(es.CreateUnit(cli))
	r.Use(middleware.Logger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})
	r.Post("/commands/demo", es.NewCommander[*commands.CreateUser]())
	r.Get("/users", userQueryFunc(cli))

	http.ListenAndServe(":3000", r)
}
