package pg

import (
	"context"
	"fmt"

	"github.com/contextcloud/eventstore/es"
	"github.com/contextcloud/goutils/xgorm"
	"github.com/contextcloud/goutils/xlog"
	gormlogger "gorm.io/gorm/logger"
)

func New(ctx context.Context, cfg es.DataConfig) (es.Conn, error) {
	if cfg.Type != "pg" {
		return nil, fmt.Errorf("invalid data provider type: %s", cfg.Type)
	}
	if cfg.Pg == nil {
		return nil, fmt.Errorf("invalid pg config")
	}

	log := xlog.Logger(ctx)

	dbops := []xgorm.Option{
		xgorm.WithLogger(log.ZapLogger(), gormlogger.Info),
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
