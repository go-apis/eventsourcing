package es

import "reflect"

// EventOptions represents the configuration options
// for the event.
type EventOptions struct {
	Name string
	Type reflect.Type
}
