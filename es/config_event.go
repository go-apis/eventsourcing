package es

import (
	"reflect"
	"strings"

	"github.com/go-apis/eventsourcing/es/utils"
)

func IsTrue(str string) bool {
	values := []string{"true", "1", "yes", "y", "t"}
	for _, v := range values {
		if strings.EqualFold(str, v) {
			return true
		}
	}
	return false
}

// EventOptions represents the configuration options
// for the event.
type EventConfig struct {
	Name    string
	Type    reflect.Type
	Aliases []string
	Publish bool
	Service string
	Factory func() (interface{}, error)
}

func NewEventConfig(thisService string, evt interface{}) *EventConfig {
	var t reflect.Type

	switch raw := evt.(type) {
	case reflect.Type:
		t = raw
	default:
		t = reflect.TypeOf(raw)
	}

	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	cfg := &EventConfig{
		Type:    t,
		Name:    t.Name(),
		Service: thisService,
		Factory: func() (interface{}, error) {
			out := reflect.New(t).Interface()
			return out, nil
		},
	}

	field, ok := t.FieldByName("BaseEvent")
	if !ok {
		return cfg
	}

	tag := field.Tag.Get("es")
	if tag == "" {
		return cfg
	}

	items := utils.SplitTag(tag)
	for _, item := range items {
		split := strings.Split(item, "=")
		part1 := split[0]
		l := len(split)

		switch {
		case part1 == "alias":
			if l == 1 {
				continue
			}
			cfg.Aliases = strings.Split(split[1], ",")
		case part1 == "publish":
			if l == 1 {
				cfg.Publish = true
				continue
			}
			cfg.Publish = IsTrue(split[1])
			continue
		case part1 == "service":
			if l == 1 {
				continue
			}
			s := split[1]
			cfg.Service = s
			continue
		case l == 1:
			cfg.Name = split[0]
			continue
		}
	}

	return cfg
}

type IsEvent interface {
	Event()
}

type BaseEvent struct {
}

func (e *BaseEvent) Event() {}
