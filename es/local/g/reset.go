package g

import (
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ResetDb(dsn string) error {
	cfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		return err
	}
	database := cfg.Database
	cfg.Database = "postgres"

	conn := stdlib.OpenDB(*cfg)

	db, err := gorm.Open(postgres.New(postgres.Config{
		Conn: conn,
	}), &gorm.Config{})

	query := `
		select pg_terminate_backend(pg_stat_activity.pid)
		from pg_stat_activity
		where pg_stat_activity.datname = ? and pid <> pg_backend_pid()`
	if err := db.Exec(query, database).Error; err != nil {
		return err
	}

	q1 := fmt.Sprintf(`drop database if exists %s`, database)
	if err := db.Exec(q1).Error; err != nil {
		return err
	}

	q2 := fmt.Sprintf(`create database %s`, database)
	if err := db.Exec(q2).Error; err != nil {
		return err
	}

	return conn.Close()
}
