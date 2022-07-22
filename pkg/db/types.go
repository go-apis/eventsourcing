package db

import (
	"strings"
	"time"

	"github.com/contextcloud/eventstore/es"
	"github.com/google/uuid"

	"github.com/jinzhu/inflection"
)

type Event struct {
	ServiceName   string    `gorm:"primaryKey"`
	Namespace     string    `gorm:"primaryKey"`
	AggregateId   uuid.UUID `gorm:"primaryKey;type:uuid"`
	AggregateType string    `gorm:"primaryKey"`
	Version       int       `gorm:"primaryKey"`
	Type          string    `gorm:"primaryKey"`
	Timestamp     time.Time
	Data          interface{} `gorm:"type:jsonb"`
	Metadata      es.Metadata `gorm:"type:jsonb"`
}

// todo add version
type Snapshot struct {
	ServiceName string `gorm:"primaryKey"`
	Namespace   string `gorm:"primaryKey"`
	Id          string `gorm:"primaryKey;type:uuid"`
	Type        string `gorm:"primaryKey"`
	Revision    string `gorm:"primaryKey"`
	Aggregate   []byte `gorm:"type:jsonb"`
}

type Entity struct {
	Namespace string `gorm:"primaryKey"`
	Id        string `gorm:"primaryKey;type:uuid"`
}

func TableName(serviceName string, aggregateName string) string {
	split := strings.Split(aggregateName, ".")
	return strings.ToLower(serviceName + "_" + inflection.Plural(split[1]))
}
