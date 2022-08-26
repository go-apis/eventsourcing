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

type Publish struct {
	EventServiceName   string    `json:"event_service_name" gorm:"primaryKey"`
	EventNamespace     string    `json:"event_namespace" gorm:"primaryKey"`
	EventAggregateId   uuid.UUID `json:"event_aggregate_id" gorm:"primaryKey;type:uuid"`
	EventAggregateType string    `json:"event_aggregate_type" gorm:"primaryKey"`
	EventVersion       int       `json:"event_version" gorm:"primaryKey"`
	EventType          string    `json:"event_type" gorm:"primaryKey"`

	ServiceName string `json:"service_name" gorm:"primaryKey"`
	Revision    string `json:"revision" gorm:"primaryKey"`

	State     string    `json:"state"`
	Timestamp time.Time `json:"timestamp"`
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
