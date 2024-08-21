package es

import (
	"reflect"
	"strings"

	"github.com/go-apis/eventsourcing/es/utils"
)

// EventOptions represents the configuration options
// for the event.
type EventHandlerConfig struct {
	Name  string
	Type  reflect.Type
	Group string
}

func NewEventHandlerConfig(h interface{}) *EventHandlerConfig {
	var t reflect.Type

	switch raw := h.(type) {
	case reflect.Type:
		t = raw
	default:
		t = reflect.TypeOf(raw)
	}

	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	cfg := &EventHandlerConfig{
		Type:  t,
		Name:  t.Name(),
		Group: ExternalGroup,
	}

	fieldNames := []string{"BaseEventHandler", "BaseSaga", "BaseProjector"}
	for _, fieldName := range fieldNames {
		field, ok := t.FieldByName(fieldName)
		if !ok {
			continue
		}

		// default project to internal group
		if fieldName == "BaseProjector" {
			cfg.Group = InternalGroup
		}

		tag := field.Tag.Get("es")
		if tag == "" {
			continue
		}

		items := utils.SplitTag(tag)
		for _, item := range items {
			split := strings.Split(item, "=")
			part1 := split[0]
			l := len(split)

			switch {
			case part1 == "group":
				if l == 1 {
					continue
				}
				cfg.Group = split[1]
			case l == 1:
				cfg.Name = split[0]
				continue
			}
		}
	}

	return cfg
}
