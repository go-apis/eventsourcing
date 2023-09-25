package pg

import (
	"context"
	"fmt"

	"github.com/contextcloud/eventstore/es"
	"github.com/contextcloud/goutils/xgorm"
)

func New(cfg es.DataConfig) (es.Conn, error) {
	if cfg.Type != "pg" {
		return nil, fmt.Errorf("invalid data provider type: %s", cfg.Type)
	}
	if cfg.Pg == nil {
		return nil, fmt.Errorf("invalid pg config")
	}

	// create a new gorm connection
	ctx := context.Background()

	dbops := []xgorm.Option{
		xgorm.WithTracing(),
	}
	if cfg.Reset {
		dbops = append(dbops, xgorm.WithRecreate())
	}

	gdb, err := xgorm.NewDb(ctx, cfg.Pg, dbops...)
	if err != nil {
		return nil, err
	}

	return NewConn(gdb)
}

func init() {
	es.RegisterDataProviders("pg", New)
}
