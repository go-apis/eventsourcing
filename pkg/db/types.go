package db

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"

	"github.com/jinzhu/inflection"
)

type Event struct {
	ServiceName   string            `json:"service_name" gorm:"primaryKey"`
	Namespace     string            `json:"namespace" gorm:"primaryKey"`
	AggregateId   uuid.UUID         `json:"aggregate_id" gorm:"primaryKey;type:uuid"`
	AggregateType string            `json:"aggregate_type" gorm:"primaryKey"`
	Version       int               `json:"version" gorm:"primaryKey"`
	Type          string            `json:"type" gorm:"primaryKey"`
	Timestamp     time.Time         `json:"timestamp"`
	Data          json.RawMessage   `json:"data" gorm:"type:jsonb"`
	Metadata      datatypes.JSONMap `json:"metadata" gorm:"type:jsonb;serializer:json"`
}

// todo add version
type Snapshot struct {
	ServiceName   string          `gorm:"primaryKey"`
	Namespace     string          `gorm:"primaryKey"`
	AggregateId   uuid.UUID       `gorm:"primaryKey;type:uuid"`
	AggregateType string          `gorm:"primaryKey"`
	Revision      string          `gorm:"primaryKey"`
	Aggregate     json.RawMessage `gorm:"type:jsonb"`
}

type Entity struct {
	Namespace string `gorm:"primaryKey"`
	Id        string `gorm:"primaryKey;type:uuid"`
}

func TableName(serviceName string, aggregateName string) string {
	return strings.ToLower(serviceName + "_" + inflection.Plural(aggregateName))
}
