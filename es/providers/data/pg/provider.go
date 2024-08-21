package pg

import (
	"context"
	"fmt"

	"github.com/go-apis/eventsourcing/es"
	"github.com/go-apis/eventsourcing/es/internal/gdb"
	"github.com/go-apis/utils/xgorm"
	"github.com/go-apis/utils/xlog"
	gormlogger "gorm.io/gorm/logger"
)

func New(ctx context.Context, cfg *es.ProviderConfig, reg es.Registry) (es.Conn, error) {
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

	db, err := xgorm.NewDb(ctx, cfg.Data.Pg, dbops...)
	if err != nil {
		return nil, err
	}

	if err := gdb.AutoMigrate(ctx, db, cfg.Service, reg); err != nil {
		return nil, err
	}

	return gdb.NewConn(ctx, cfg.Service, db, reg, false)
}

func init() {
	es.RegisterDataProviders("pg", New)
}
