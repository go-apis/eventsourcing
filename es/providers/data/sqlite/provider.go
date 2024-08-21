package pg

import (
	"context"
	"fmt"

	"github.com/go-apis/eventsourcing/es"
	"github.com/go-apis/eventsourcing/es/internal/gdb"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func New(ctx context.Context, cfg *es.ProviderConfig, reg es.Registry) (es.Conn, error) {
	if cfg.Data.Type != "sqlite" {
		return nil, fmt.Errorf("invalid data provider type: %s", cfg.Data.Type)
	}
	if cfg.Data.Sqlite == nil {
		return nil, fmt.Errorf("invalid sqlite config")
	}

	dsn := cfg.Data.Sqlite.File
	if cfg.Data.Sqlite.Memory {
		dsn = ":memory:"
	}
	if dsn == "" {
		return nil, fmt.Errorf("invalid sqlite file")
	}

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := gdb.AutoMigrate(ctx, db, cfg.Service, reg); err != nil {
		return nil, err
	}

	return gdb.NewConn(ctx, cfg.Service, db, reg, true)
}

func init() {
	es.RegisterDataProviders("sqlite", New)
}
