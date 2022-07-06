package local

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ResetDb(opts ...OptionFunc) error {
	o := NewOptions()
	for _, opt := range opts {
		opt(o)
	}
	dbName := o.DbName
	WithDbName("postgres")(o)

	dsn := o.DSN()

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		return err
	}

	query := `
		select pg_terminate_backend(pg_stat_activity.pid)
		from pg_stat_activity
		where pg_stat_activity.datname = ? and pid <> pg_backend_pid()`
	if err := db.Exec(query, dbName).Error; err != nil {
		return err
	}

	q1 := fmt.Sprintf(`drop database if exists %s`, dbName)
	if err := db.Exec(q1).Error; err != nil {
		return err
	}

	q2 := fmt.Sprintf(`create database %s`, dbName)
	if err := db.Exec(q2).Error; err != nil {
		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
