package es

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

var (
	// ErrEventHandlerAlreadySet is when a handler is already registered for a command.
	ErrEventHandlerAlreadySet = errors.New("handler is already set")
	// ErrEventHandlerNotFound is when no handler can be found.
	ErrEventHandlerNotFound = errors.New("no handlers for command")
)

type EventRegistry interface {
	GroupEventHandler

	AddEvent(eventConfig *EventConfig) error
	AddGroupEventHandler(h EventHandler, group string, eventConfig *EventConfig) error

	GetGroups() []string
	GetEventConfig(service string, eventType string) (*EventConfig, error)
	ParseEvent(ctx context.Context, msg []byte) (*Event, error)
}

type eventRegistry struct {
	hash  map[string]*EventConfig
	typed map[reflect.Type]*EventConfig

	groupHash     map[string]bool
	groups        []string
	groupHandlers map[string]EventHandlers
}

func (r *eventRegistry) HandleGroupEvent(ctx context.Context, group string, evt *Event) error {
	key := group + strings.ToLower("__"+evt.Service+"__"+evt.Type)
	handlers, ok := r.groupHandlers[key]
	if !ok {
		return nil
	}

	withNs := SetNamespace(ctx, evt.Namespace)
	return handlers.Handle(withNs, evt)
}

func (r *eventRegistry) AddEvent(eventConfig *EventConfig) error {
	if _, ok := r.typed[eventConfig.Type]; ok {
		// already registered.
		return nil
	}

	name := strings.ToLower(eventConfig.Service + "__" + eventConfig.Name)
	if _, ok := r.hash[name]; ok {
		return fmt.Errorf("duplicate event: %s", name)
	}
	r.hash[name] = eventConfig

	for _, alias := range eventConfig.Aliases {
		aliasName := strings.ToLower(eventConfig.Service + "__" + alias)
		if aliasName == name {
			continue
		}
		if _, ok := r.hash[aliasName]; ok {
			return fmt.Errorf("duplicate event: %s - %s", name, aliasName)
		}
		r.hash[aliasName] = eventConfig
	}

	r.typed[eventConfig.Type] = eventConfig
	return nil
}
func (r *eventRegistry) GetGroups() []string {
	return r.groups
}
func (r *eventRegistry) GetEventConfig(service string, eventType string) (*EventConfig, error) {
	name := strings.ToLower(service + "__" + eventType)
	if evt, ok := r.hash[name]; ok {
		return evt, nil
	}
	return nil, fmt.Errorf("event %s: %w", name, ErrNotFound)
}
func (r *eventRegistry) ParseEvent(ctx context.Context, msg []byte) (*Event, error) {
	out := struct {
		*Event
		Data json.RawMessage `json:"data"`
	}{}
	if err := json.Unmarshal(msg, &out); err != nil {
		return nil, fmt.Errorf("could not decode event: %w", err)
	}

	evtConfig, err := r.GetEventConfig(out.Service, out.Type)
	if err != nil {
		return nil, err
	}

	data, err := evtConfig.Factory()
	if err != nil {
		return nil, fmt.Errorf("could not create event: %w", err)
	}

	if err := json.Unmarshal(out.Data, data); err != nil {
		return nil, fmt.Errorf("could not decode event: %w", err)
	}

	metadata := out.Metadata
	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	return &Event{
		Service:       out.Service,
		Namespace:     out.Namespace,
		AggregateId:   out.AggregateId,
		AggregateType: out.AggregateType,
		Version:       out.Version,
		Type:          out.Type,
		By:            out.By,
		Timestamp:     out.Timestamp,
		Metadata:      metadata,
		Data:          data,
	}, nil
}
func (r *eventRegistry) AddGroupEventHandler(h EventHandler, group string, eventConfig *EventConfig) error {
	if _, ok := r.groupHash[group]; !ok {
		r.groupHash[group] = true
		r.groups = append(r.groups, group)
	}

	key := group + strings.ToLower("__"+eventConfig.Service+"__"+eventConfig.Name)
	r.groupHandlers[key] = append(r.groupHandlers[key], h)

	return r.AddEvent(eventConfig)
}

func NewEventRegistry() EventRegistry {
	return &eventRegistry{
		hash:          make(map[string]*EventConfig),
		typed:         make(map[reflect.Type]*EventConfig),
		groupHash:     make(map[string]bool),
		groups:        []string{},
		groupHandlers: make(map[string]EventHandlers),
	}
}
