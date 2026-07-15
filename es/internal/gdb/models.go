package gdb

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
	ServiceName   string            `json:"service_name" gorm:"primaryKey" dynmgrm:"pk"`
	Namespace     string            `json:"namespace" gorm:"primaryKey" dynmgrm:"sk"`
	AggregateId   uuid.UUID         `json:"aggregate_id" gorm:"primaryKey;type:uuid" dynmgrm:"sk"`
	AggregateType string            `json:"aggregate_type" gorm:"primaryKey" dynmgrm:"sk"`
	Version       int               `json:"version" gorm:"primaryKey" dynmgrm:"sk"`
	Type          string            `json:"type" gorm:"primaryKey" dynmgrm:"sk"`
	By            *es.Actor         `json:"by" gorm:"type:jsonb;serializer:json"`
	Timestamp     time.Time         `json:"timestamp"`
	Data          json.RawMessage   `json:"data" gorm:"type:jsonb"`
	Metadata      datatypes.JSONMap `json:"metadata" gorm:"type:jsonb;serializer:json"`
}

// Outbox rows are publishable events awaiting relay to the stream. They are
// inserted in the same transaction as the events they mirror and deleted
// once published, so the table's steady state is empty; depth and row age
// are the publish-backlog health signals.
//
// Like the events table, one outbox is shared by every service on the
// database: rows are stamped with the owning service and each service's
// relay claims only its own, ordered by id via the composite index — which
// is what keeps claims cheap when a service builds a real backlog.
type Outbox struct {
	Id          int64           `json:"id" gorm:"primaryKey;autoIncrement;index:idx_outbox_service_id,priority:2"`
	ServiceName string          `json:"service_name" gorm:"index:idx_outbox_service_id,priority:1"`
	OrderingKey string          `json:"ordering_key"`
	Payload     json.RawMessage `json:"payload" gorm:"type:jsonb"`
	CreatedAt   time.Time       `json:"created_at"`
}

// TableName keeps the singular pattern name rather than gorm's "outboxes".
func (Outbox) TableName() string {
	return "outbox"
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
