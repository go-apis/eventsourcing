package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/contextcloud/eventstore/es/gstream"
	"github.com/contextcloud/eventstore/es/local"
	"github.com/contextcloud/eventstore/examples/groups/commands"
	"github.com/contextcloud/eventstore/examples/groups/config"
	"github.com/contextcloud/eventstore/pkg/db"
	"github.com/contextcloud/eventstore/pkg/pub"
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

func setupTracer() func(context.Context) error {
	url := "http://localhost:9411/api/v2/spans"

	exporter, err := zipkin.New(url)
	if err != nil {
		panic(err)
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("groups"),
		)),
	)
	otel.SetTracerProvider(tp)
	return tp.Shutdown
}

func main() {
	shutdown := setupTracer()
	defer shutdown(context.Background())

	cfg, err := config.EventStoreConfig()
	if err != nil {
		panic(err)
	}

	conn, err := local.NewConn(
		db.WithDbUser("es"),
		db.WithDbPassword("es"),
		db.WithDbName("eventstore"),
	)
	if err != nil {
		panic(err)
	}

	streamer, err := gstream.NewStreamer(
		pub.WithProjectId("nordic-gaming"),
		pub.WithTopicId("contextcloud_example"),
	)
	if err != nil {
		panic(err)
	}

	cli, err := es.NewClient(cfg, conn, streamer)
	if err != nil {
		panic(err)
	}

	serverCtx, serverStopCtx := context.WithCancel(context.Background())
	if err := cli.Initialize(serverCtx); err != nil {
		panic(err)
	}

	r := chi.NewRouter()
	r.Use(otelchi.Middleware("server", otelchi.WithChiRoutes(r)))
	r.Use(es.CreateUnit(cli))
	r.Use(middleware.Logger)
	r.Post("/commands/creategroup", es.NewCommander[*commands.CreateGroup]())

	// The HTTP Server
	server := &http.Server{Addr: ":3001", Handler: r}

	// Listen for syscall signals for process to interrupt/quit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-sig

		// Shutdown signal with grace period of 30 seconds
		shutdownCtx, _ := context.WithTimeout(serverCtx, 30*time.Second)

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				log.Fatal("graceful shutdown timed out.. forcing exit.")
			}
		}()

		// Trigger graceful shutdown
		err := server.Shutdown(shutdownCtx)
		if err != nil {
			log.Fatal(err)
		}
		serverStopCtx()
	}()

	// Run the server
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}

	// Wait for server context to be stopped
	<-serverCtx.Done()
}
