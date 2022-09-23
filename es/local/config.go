package local

import (
	"github.com/contextcloud/eventstore/es"
	"github.com/contextcloud/eventstore/pkg/db"
)

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
}

func NewConn(cfg *Config) (es.Conn, error) {
	opts := []db.OptionFunc{
		db.WithDbHost(cfg.Host),
		db.WithDbPort(cfg.Port),
		db.WithDbUser(cfg.User),
		db.WithDbPassword(cfg.Password),
		db.WithDbName(cfg.Name),
	}

	gormDb, err := db.Open(opts...)
	if err != nil {
		return nil, err
	}

	c := &conn{
		db: gormDb,
	}
	return c, nil
}
