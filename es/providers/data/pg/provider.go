package pg

import (
	"context"
	"fmt"

	"github.com/contextcloud/eventstore/es"
	"github.com/contextcloud/goutils/xgorm"
	"github.com/contextcloud/goutils/xlog"
	gormlogger "gorm.io/gorm/logger"
)

func New(ctx context.Context, cfg *es.ProviderConfig) (es.Conn, error) {
	if cfg.Data.Type != "pg" {
		return nil, fmt.Errorf("invalid data provider type: %s", cfg.Data.Type)
	}
	if cfg.Data.Pg == nil {
		return nil, fmt.Errorf("invalid pg config")
	}

	log := xlog.Logger(ctx)

	dbops := []xgorm.Option{
		xgorm.WithLogger(log.ZapLogger(), gormlogger.Info),
		xgorm.WithTracing(),
		xgorm.WithDisableNestedTransaction(),
		xgorm.WithSkipDefaultTransaction(),
	}
	if cfg.Data.Reset {
		dbops = append(dbops, xgorm.WithRecreate())
	}

	gdb, err := xgorm.NewDb(ctx, cfg.Data.Pg, dbops...)
	if err != nil {
		return nil, err
	}

	return NewConn(ctx, cfg.Service, gdb)
}

func init() {
	es.RegisterDataProviders("pg", New)
}
