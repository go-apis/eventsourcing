package pg

import (
	"eventstore/es"
	"time"

	"github.com/uptrace/bun"
)

type dbEvent struct {
	bun.BaseModel `bun:"table:events,alias:evt"`

	ServiceName   string `bun:",pk"`
	Namespace     string `bun:",pk"`
	AggregateId   string `bun:",pk,type:uuid"`
	AggregateType string `bun:",pk"`
	Version       int    `bun:",pk"`
	Type          string `bun:",notnull"`
	Timestamp     time.Time
	Data          []byte      `bun:"type:jsonb"`
	Metadata      es.Metadata `bun:"type:jsonb"`
}

// todo add version
type dbSnapshot struct {
	bun.BaseModel `bun:"table:events,alias:evt"`

	ServiceName string `bun:",pk"`
	Namespace   string `bun:",pk"`
	Id          string `bun:",pk,type:uuid"`
	Type        string `bun:",pk"`
	Revision    string `bun:",pk"`
	Aggregate   []byte `bun:",notnull,type:jsonb"`
}
