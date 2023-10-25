package es

import (
	"reflect"
)

// EventOptions represents the configuration options
// for the event.
type EventConfig struct {
	Name    string
	Type    reflect.Type
	Publish bool
	Factory func() (interface{}, error)
}

func NewEventConfig(evt interface{}) *EventConfig {
	var t reflect.Type

	switch raw := evt.(type) {
	case reflect.Type:
		t = raw
		break
	default:
		t = reflect.TypeOf(raw)
		break
	}

	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	name := t.Name()
	factory := func() (interface{}, error) {
		out := reflect.New(t).Interface()
		return out, nil
	}

	impl, _ := factory()
	_, publish := impl.(EventPublish)

	return &EventConfig{
		Name:    name,
		Type:    t,
		Publish: publish,
		Factory: factory,
	}
}

type EventPublish interface {
	Publish()
}

type BaseEventPublish struct {
}

func (b BaseEventPublish) Publish() {
}

type EventPublished interface {
	Published()
}

type BaseEventPublished struct {
}

func (b BaseEventPublished) Published() {
}
