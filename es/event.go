package es

import (
	"fmt"
	"reflect"
	"time"

	"github.com/google/uuid"
)

// EventDataMapper for creating event data from a given event name
type EventDataMapper map[string]EventDataFunc

func NewEventDataFunc(objs ...interface{}) EventDataMapper {
	m := make(EventDataMapper)

	for _, obj := range objs {
		t := reflect.TypeOf(obj)
		for t.Kind() == reflect.Ptr {
			t = t.Elem()
		}

		m[t.Name()] = func() (interface{}, error) {
			out := reflect.New(t).Interface()
			return out, nil
		}
	}

	return m
}

// EventDataFunc for creating a Data
type EventDataFunc func() (interface{}, error)

// Event that has been persisted to the event store.
type Event struct {
	ServiceName   string                 `json:"service_name"`
	Namespace     string                 `json:"namespace"`
	AggregateId   uuid.UUID              `json:"aggregate_id"`
	AggregateType string                 `json:"aggregate_type"`
	Version       int                    `json:"version"`
	Type          string                 `json:"type"`
	Timestamp     time.Time              `json:"timestamp"`
	Data          interface{}            `json:"data"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// String implements the String method of the Event interface.
func (e Event) String() string {
	return fmt.Sprintf("%s@%d", e.Type, e.Version)
}
