package pg

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
)

type PostgresFactory interface {
	New() (*bun.DB, error)
}

type pgFactory struct {
	dsn     string
	logging bool
}

func (s *pgFactory) New() (*bun.DB, error) {
	conn := pgdriver.NewConnector(
		pgdriver.WithDSN(s.dsn),
	)
	sqldb := sql.OpenDB(conn)
	db := bun.NewDB(sqldb, pgdialect.New(), bun.WithDiscardUnknownColumns())

	if s.logging {
		db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))
	}

	// TODO should we try to connect?
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func (s *pgFactory) recreate() error {
	cfg := &pgdriver.Config{}
	pgdriver.WithDSN(s.dsn)(cfg)
	database := cfg.Database

	conn := pgdriver.NewConnector(
		pgdriver.WithDSN(s.dsn),
		pgdriver.WithDatabase("postgres"),
	)
	sqldb := sql.OpenDB(conn)
	db := bun.NewDB(sqldb, pgdialect.New())

	query := `
		select pg_terminate_backend(pg_stat_activity.pid)
		from pg_stat_activity
		where pg_stat_activity.datname = ? and pid <> pg_backend_pid()`
	if _, err := db.Exec(query, database); err != nil {
		return err
	}

	q1 := fmt.Sprintf(`drop database if exists %s`, database)
	if _, err := db.Exec(q1); err != nil {
		return err
	}

	q2 := fmt.Sprintf(`create database %s`, database)
	if _, err := db.Exec(q2); err != nil {
		return err
	}

	return nil
}

func (s *pgFactory) createTables(types []interface{}) error {
	db, err := s.New()
	if err != nil {
		return err
	}

	ctx := context.Background()
	if _, err := db.NewCreateTable().
		Model(&types).
		IfNotExists().
		Exec(ctx); err != nil {
		return err
	}
	return nil
}

func newFactory(dsn string, logging bool, recreate bool, tables []interface{}) (PostgresFactory, error) {
	factory := &pgFactory{
		dsn:     dsn,
		logging: logging,
	}

	if recreate {
		if err := factory.recreate(); err != nil {
			return nil, err
		}
	}

	if len(tables) > 0 {
		if err := factory.createTables(tables); err != nil {
			return nil, err
		}
	}

	return factory, nil
}
