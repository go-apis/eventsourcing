package data

import "github.com/contextcloud/eventstore/pkg/db"

type Config struct {
	Type string
	Pg   *db.Config
}
