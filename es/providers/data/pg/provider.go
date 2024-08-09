package pg

import (
	"context"
	"fmt"

	"github.com/go-apis/eventsourcing/es"
)

func New(ctx context.Context, cfg *es.ProviderConfig, reg es.Registry) (es.Conn, error) {
	if cfg.Data.Type != "pg" {
		return nil, fmt.Errorf("invalid data provider type: %s", cfg.Data.Type)
	}
	if cfg.Data.Pg == nil {
		return nil, fmt.Errorf("invalid pg config")
	}
	return NewConn(ctx, cfg.Service, cfg.Data.Pg, cfg.Data.Reset, reg)
}

func init() {
	es.RegisterDataProviders("pg", New)
}
