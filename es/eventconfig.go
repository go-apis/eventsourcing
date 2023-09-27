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
	Service *string
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

	var service *string
	if field, ok := t.FieldByName("BaseEventPublished"); ok {
		if tag := field.Tag.Get("service"); tag != "" {
			service = &tag
		}

	}

	return &EventConfig{
		Name:    name,
		Type:    t,
		Publish: publish,
		Factory: factory,
		Service: service,
	}
}

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
