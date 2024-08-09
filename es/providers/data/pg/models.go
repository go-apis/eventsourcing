package pg

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/go-apis/eventsourcing/es"
	"github.com/google/uuid"
	"github.com/jinzhu/inflection"
	"gorm.io/datatypes"
)

type Notification struct {
	PID     uint32
	Channel string
	Payload string
}

type Event struct {
	ServiceName   string            `json:"service_name" gorm:"primaryKey"`
	Namespace     string            `json:"namespace" gorm:"primaryKey"`
	AggregateId   uuid.UUID         `json:"aggregate_id" gorm:"primaryKey;type:uuid"`
	AggregateType string            `json:"aggregate_type" gorm:"primaryKey"`
	Version       int               `json:"version" gorm:"primaryKey"`
	Type          string            `json:"type" gorm:"primaryKey"`
	By            *es.Actor         `json:"by" gorm:"type:jsonb;serializer:json"`
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

type PersistedCommand struct {
	ServiceName  string          `json:"service_name" gorm:"primaryKey"`
	Namespace    string          `json:"namespace" gorm:"primaryKey"`
	Id           uuid.UUID       `json:"id" gorm:"primaryKey;type:uuid"`
	Type         string          `json:"type"`
	Data         json.RawMessage `json:"data" gorm:"type:jsonb"`
	ExecuteAfter time.Time       `json:"execute_after"`
	CreatedAt    time.Time       `json:"created_at"`
	By           *es.Actor       `json:"by" gorm:"type:jsonb;serializer:json"`
}

func TableName(service string, aggregateName string) string {
	return strings.ToLower(service + "_" + inflection.Plural(aggregateName))
}
