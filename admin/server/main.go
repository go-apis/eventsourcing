package main

import (
	"context"

	api "github.com/contextcloud/eventstore/admin"
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

	h, err := api.NewHandler(serverCtx, cfg)
	if err != nil {
		panic(err)
	}

	tracer, err := srv.NewTracer(serverCtx, cfg)
	if err != nil {
		panic(err)
	}

	server, err := srv.NewStartable(cfg.SrvAddr, h)
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
