package local

import (
	"strings"
	"time"

	"github.com/contextcloud/eventstore/es"

	"github.com/jinzhu/inflection"
)

type event struct {
	ServiceName   string `gorm:"primaryKey"`
	Namespace     string `gorm:"primaryKey"`
	AggregateId   string `gorm:"primaryKey;type:uuid"`
	AggregateType string `gorm:"primaryKey"`
	Version       int    `gorm:"primaryKey"`
	Type          string `gorm:"primaryKey"`
	Timestamp     time.Time
	Data          interface{} `gorm:"type:jsonb"`
	Metadata      es.Metadata `gorm:"type:jsonb"`
}

// todo add version
type snapshot struct {
	ServiceName string `gorm:"primaryKey"`
	Namespace   string `gorm:"primaryKey"`
	Id          string `gorm:"primaryKey;type:uuid"`
	Type        string `gorm:"primaryKey"`
	Revision    string `gorm:"primaryKey"`
	Aggregate   []byte `gorm:"type:jsonb"`
}

type entity struct {
	Namespace string `gorm:"primaryKey"`
	Id        string `gorm:"primaryKey;type:uuid"`
}

func tableName(serviceName string, aggregateName string) string {
	split := strings.Split(aggregateName, ".")
	return strings.ToLower(serviceName + "_" + inflection.Plural(split[1]))
}
