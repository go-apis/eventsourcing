package db

import (
	"fmt"
)

func Reset(cfg *Config) error {
	r := &Config{
		Host:     cfg.Host,
		Port:     cfg.Port,
		User:     cfg.User,
		Password: cfg.Password,
		Name:     "postgres",
		Debug:    cfg.Debug,
	}

	db, err := Open(r)
	if err != nil {
		return err
	}

	query := `
		select pg_terminate_backend(pg_stat_activity.pid)
		from pg_stat_activity
		where pg_stat_activity.datname = ? and pid <> pg_backend_pid()`
	if err := db.Exec(query, cfg.Name).Error; err != nil {
		return err
	}

	q1 := fmt.Sprintf(`drop database if exists %s`, cfg.Name)
	if err := db.Exec(q1).Error; err != nil {
		return err
	}

	q2 := fmt.Sprintf(`create database %s`, cfg.Name)
	if err := db.Exec(q2).Error; err != nil {
		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
