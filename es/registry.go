package es

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/contextcloud/eventstore/es/utils"
)

var GlobalRegistry = NewRegistry()

type Registry struct {
	entities             map[string]*EntityConfig
	commands             map[string]*CommandConfig
	eventsByType         map[reflect.Type]*EventConfig
	events               map[string]*EventConfig
	commandHandlers      map[reflect.Type]CommandHandler
	groupedEventHandlers map[string]map[reflect.Type]EventHandlers
	eventHandlerGroups   map[string]bool
	replayHandlers       map[string]CommandHandler
}

func (r *Registry) AddEntity(entityConfig *EntityConfig) error {
	name := strings.ToLower(entityConfig.Name)
	if _, ok := r.entities[name]; ok {
		return fmt.Errorf("duplicate entity: %s", name)
	}
	r.entities[name] = entityConfig
	return nil
}
func (r *Registry) AddCommandHandler(commandConfig *CommandConfig, h CommandHandler) error {
	name := strings.ToLower(commandConfig.Name)

	if _, ok := r.commands[name]; ok {
		return fmt.Errorf("duplicate command: %s", name)
	}
	if _, ok := r.commandHandlers[commandConfig.Type]; ok {
		return fmt.Errorf("duplicate command handler: %s", commandConfig.Type)
	}

	r.commandHandlers[commandConfig.Type] = h
	r.commands[name] = commandConfig
	return nil
}
func (r *Registry) AddEventHandler(group string, eventConfig *EventConfig, h EventHandler) error {
	r.eventHandlerGroups[group] = true

	grouped, ok := r.groupedEventHandlers[group]
	if !ok {
		r.groupedEventHandlers[group] = map[reflect.Type]EventHandlers{
			eventConfig.Type: {h},
		}
	} else {
		r.groupedEventHandlers[group][eventConfig.Type] = append(grouped[eventConfig.Type], h)
	}
	return r.AddEvent(eventConfig)
}
func (r *Registry) AddReplayHandler(entityConfig *EntityConfig, h CommandHandler) error {
	name := strings.ToLower(entityConfig.Name)
	if _, ok := r.replayHandlers[name]; ok {
		return fmt.Errorf("duplicate replay handler: %s", name)
	}
	r.replayHandlers[name] = h
	return nil
}
func (r *Registry) AddEvent(eventConfig *EventConfig) error {
	if _, ok := r.eventsByType[eventConfig.Type]; ok {
		// already registered.
		return nil
	}

	name := strings.ToLower(eventConfig.Service + "__" + eventConfig.Name)
	if _, ok := r.events[name]; ok {
		return fmt.Errorf("duplicate event: %s", name)
	}
	r.events[name] = eventConfig

	for _, alias := range eventConfig.Aliases {
		aliasName := strings.ToLower(eventConfig.Service + "__" + alias)
		if aliasName == name {
			continue
		}
		if _, ok := r.events[aliasName]; ok {
			return fmt.Errorf("duplicate event: %s - %s", name, aliasName)
		}
		r.events[aliasName] = eventConfig
	}

	r.eventsByType[eventConfig.Type] = eventConfig
	return nil
}

func (r *Registry) GetEventConfig(serviceName string, eventType string) (*EventConfig, error) {
	name := strings.ToLower(serviceName + "__" + eventType)
	if evt, ok := r.events[name]; ok {
		return evt, nil
	}
	return nil, fmt.Errorf("event %s: %w", name, ErrNotFound)
}
func (r *Registry) GetEntityConfig(entityConfig string) (*EntityConfig, error) {
	name := strings.ToLower(entityConfig)
	if entity, ok := r.entities[name]; ok {
		return entity, nil
	}
	return nil, fmt.Errorf("entity %s: %w", name, ErrNotFound)
}
func (r *Registry) GetCommandHandler(command interface{}) (CommandHandler, error) {
	t := utils.GetElemType(command)
	if h, ok := r.commandHandlers[t]; ok {
		return h, nil
	}
	return nil, fmt.Errorf("command handler %v: %w", t, ErrNotFound)
}
func (r *Registry) GetReplayHandler(command *ReplayCommand) (CommandHandler, error) {
	n := strings.ToLower(command.AggregateName)
	if h, ok := r.replayHandlers[n]; ok {
		return h, nil
	}
	return nil, fmt.Errorf("replay handler %s: %w", n, ErrNotFound)
}
func (r *Registry) GetEventHandlers(group string, event interface{}) EventHandlers {
	grouped, ok := r.groupedEventHandlers[group]
	if !ok {
		return nil
	}

	t := utils.GetElemType(event)
	return grouped[t]
}
func (r *Registry) GetEventHandlerGroups() []string {
	var out []string
	for k := range r.eventHandlerGroups {
		out = append(out, k)
	}
	return out
}
func (r *Registry) GetEntities() []*EntityConfig {
	var out []*EntityConfig
	for _, v := range r.entities {
		out = append(out, v)
	}
	return out
}

func NewRegistry() *Registry {
	return &Registry{
		entities:             map[string]*EntityConfig{},
		commands:             map[string]*CommandConfig{},
		events:               map[string]*EventConfig{},
		eventsByType:         map[reflect.Type]*EventConfig{},
		commandHandlers:      map[reflect.Type]CommandHandler{},
		groupedEventHandlers: map[string]map[reflect.Type]EventHandlers{},
		eventHandlerGroups:   map[string]bool{},
		replayHandlers:       map[string]CommandHandler{},
	}
}

func RegistryAdd(service string, items ...interface{}) error {
	var sagas []IsSaga
	var projectors []IsProjector
	var eventHandlers []IsEventHandler
	var commandHandlers []IsCommandHandler
	var aggregates []Aggregate
	var entities []Entity
	var events []interface{}

	for _, item := range items {
		switch raw := item.(type) {
		case IsSaga:
			sagas = append(sagas, raw)
			continue
		case IsProjector:
			projectors = append(projectors, raw)
			continue
		case IsEventHandler:
			eventHandlers = append(eventHandlers, raw)
			continue
		case IsCommandHandler:
			commandHandlers = append(commandHandlers, raw)
			continue
		case IsEvent:
			events = append(events, raw)

		case Aggregate:
			aggregates = append(aggregates, raw)
			continue
		case Entity:
			entities = append(entities, raw)
			continue
		default:
			return fmt.Errorf("invalid item type: %T", item)
		}
	}

	// register entities
	for _, entity := range entities {
		opts := NewEntityOptions(entity)
		entityConfig, err := NewEntityConfig(opts)
		if err != nil {
			return err
		}
		if err := GlobalRegistry.AddEntity(entityConfig); err != nil {
			return err
		}
	}

	// dynamic aggregates
	for _, agg := range aggregates {
		opts := NewEntityOptions(agg)
		entityConfig, err := NewEntityConfig(opts)
		if err != nil {
			return err
		}
		// handles!
		commandHandles := NewCommandHandles(agg)
		h := NewAggregateHandler(entityConfig, commandHandles)

		if err := GlobalRegistry.AddEntity(entityConfig); err != nil {
			return err
		}
		if err := GlobalRegistry.AddReplayHandler(entityConfig, h); err != nil {
			return err
		}
		for t := range commandHandles {
			commandConfig := NewCommandConfig(t)
			if err := GlobalRegistry.AddCommandHandler(commandConfig, h); err != nil {
				return err
			}
		}
	}

	// handlers
	for _, commandHandler := range commandHandlers {
		commandHandles := NewCommandHandles(commandHandler)
		h := NewCommandHandler(commandHandler, commandHandles)

		for t := range commandHandles {
			commandHandler := NewCommandConfig(t)
			if err := GlobalRegistry.AddCommandHandler(commandHandler, h); err != nil {
				return err
			}
		}
	}

	// sagas
	for _, saga := range sagas {
		eventHandlerConfig := NewEventHandlerConfig(saga)
		handles := NewSagaHandles(saga)
		h := NewSagaEventHandler(handles, saga)

		for t := range handles {
			eventConfig := NewEventConfig(service, t)
			if err := GlobalRegistry.AddEventHandler(eventHandlerConfig.Group, eventConfig, h); err != nil {
				return err
			}
		}
	}

	// projectors
	for _, projector := range projectors {
		eventHandlerConfig := NewEventHandlerConfig(projector)
		handles := FindProjectorHandles(projector)

		matrix := map[reflect.Type]map[reflect.Type][]*ProjectorHandle{}
		for _, h := range handles {
			if _, ok := matrix[h.AggregateType]; !ok {
				matrix[h.AggregateType] = map[reflect.Type][]*ProjectorHandle{}
			}
			matrix[h.AggregateType][h.EventType] = append(matrix[h.AggregateType][h.EventType], h)
		}

		for agg, m := range matrix {
			opts := NewEntityOptions(agg)
			entityConfig, err := NewEntityConfig(opts)
			if err != nil {
				return err
			}
			for evt, handles := range m {
				eventConfig := NewEventConfig(service, evt)
				h := NewProjectorEventHandler(entityConfig, handles, projector)

				if err := GlobalRegistry.AddEventHandler(eventHandlerConfig.Group, eventConfig, h); err != nil {
					return err
				}
			}
		}
	}

	// eventhandlers
	for _, eventHandler := range eventHandlers {
		eventHandlerConfig := NewEventHandlerConfig(eventHandler)
		handles := NewEventHandles(eventHandler)
		h := NewEventHandler(eventHandler, handles)

		for t := range handles {
			eventConfig := NewEventConfig(service, t)
			if err := GlobalRegistry.AddEventHandler(eventHandlerConfig.Group, eventConfig, h); err != nil {
				return err
			}
		}
	}

	// events
	for _, evt := range events {
		evtConfig := NewEventConfig(service, evt)
		if err := GlobalRegistry.AddEvent(evtConfig); err != nil {
			return err
		}
	}

	return nil
}
