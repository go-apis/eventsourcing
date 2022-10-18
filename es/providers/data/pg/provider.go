package pg

import (
	"fmt"

	"github.com/contextcloud/eventstore/es"
	"github.com/contextcloud/eventstore/pkg/pgdb"
)

func New(cfg es.DataConfig) (es.Conn, error) {
	if cfg.Type != "pg" {
		return nil, fmt.Errorf("invalid data provider type: %s", cfg.Type)
	}
	if cfg.Pg == nil {
		return nil, fmt.Errorf("invalid pg config")
	}

	// create a new gorm connection
	gdb, err := pgdb.Open(cfg.Pg)
	if err != nil {
		return nil, err
	}

	return NewConn(gdb)
}

func init() {
	es.RegisterDataProviders("pg", New)
}
