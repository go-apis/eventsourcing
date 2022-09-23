package main

import (
	"context"

	"github.com/contextcloud/eventstore/server"
	"github.com/contextcloud/graceful"
	"github.com/contextcloud/graceful/config"
	"github.com/contextcloud/graceful/srv"
)

func main() {
	serverCtx, serverCancel := context.WithCancel(context.Background())

	cfg, err := config.NewConfig(serverCtx)
	if err != nil {
		panic(err)
	}

	tracer, err := srv.NewTracer(serverCtx, cfg)
	if err != nil {
		panic(err)
	}

	server, err := server.NewServer(serverCtx, cfg)
	if err != nil {
		panic(err)
	}

	health := srv.NewHealth(cfg.HealthAddr)

	multi := srv.NewMulti(
		server,
		health,
		tracer,
	)
	graceful.Run(serverCtx, multi)
	serverCancel()
}
