package es

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/contextcloud/eventstore/es/utils"
)

var GlobalRegistry = Registry{
	entities:        map[string]*EntityConfig{},
	commands:        map[string]*CommandConfig{},
	events:          map[string]*EventConfig{},
	commandHandlers: map[reflect.Type]CommandHandler{},
	eventHandlers:   map[reflect.Type][]EventHandler{},
	replayHandlers:  map[string]CommandHandler{},
}

type Registry struct {
	entities        map[string]*EntityConfig
	commands        map[string]*CommandConfig
	events          map[string]*EventConfig
	commandHandlers map[reflect.Type]CommandHandler
	eventHandlers   map[reflect.Type][]EventHandler
	replayHandlers  map[string]CommandHandler
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
func (r *Registry) AddEventHandler(eventConfig *EventConfig, h EventHandler) error {
	name := strings.ToLower(eventConfig.Name)
	if _, ok := r.events[name]; !ok {
		r.events[name] = eventConfig
	}
	r.eventHandlers[eventConfig.Type] = append(r.eventHandlers[eventConfig.Type], h)
	return nil
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
	name := strings.ToLower(eventConfig.Name)
	if _, ok := r.events[name]; !ok {
		r.events[name] = eventConfig
	}
	return nil
}

func (r *Registry) GetEventConfig(eventType string) (*EventConfig, error) {
	name := strings.ToLower(eventType)
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
func (r *Registry) GetReplayHandler(command interface{}) (CommandHandler, error) {
	t := utils.GetElemType(command)
	n := strings.ToLower(t.Name())
	if h, ok := r.replayHandlers[n]; ok {
		return h, nil
	}
	return nil, fmt.Errorf("command handler %v: %w", t, ErrNotFound)
}
func (r *Registry) GetEventHandlers(event interface{}) []EventHandler {
	t := utils.GetElemType(event)
	return r.eventHandlers[t]
}
func (r *Registry) GetEntities() []*EntityConfig {
	var out []*EntityConfig
	for _, v := range r.entities {
		out = append(out, v)
	}
	return out
}

func RegistryAdd(items ...interface{}) error {
	var handlers []IsCommandHandler
	var sagas []IsSaga
	var projectors []IsProjector
	var aggregates []Aggregate
	var entities []Entity
	var aggregateConfigs []*AggregateConfig
	var middlewares []CommandHandlerMiddleware
	var events []interface{}

	for _, item := range items {
		switch raw := item.(type) {
		case IsCommandHandler:
			handlers = append(handlers, raw)
			continue
		case IsSaga:
			sagas = append(sagas, raw)
			continue
		case IsProjector:
			projectors = append(projectors, raw)
			continue
		case Aggregate:
			aggregates = append(aggregates, raw)
			continue
		case Entity:
			entities = append(entities, raw)
			continue
		case *AggregateConfig:
			aggregateConfigs = append(aggregateConfigs, raw)
			continue
		case CommandHandlerMiddleware:
			middlewares = append(middlewares, raw)
			continue
		case EventPublish:
			events = append(events, raw)
		case EventPublished:
			events = append(events, raw)
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

	// basic aggregates
	for _, cfg := range aggregateConfigs {
		entityConfig, err := NewEntityConfig(cfg.EntityOptions)
		if err != nil {
			return err
		}
		h := NewAggregateHandler(entityConfig, nil)
		h = UseCommandHandlerMiddleware(h, middlewares...)

		if err := GlobalRegistry.AddEntity(entityConfig); err != nil {
			return err
		}
		if err := GlobalRegistry.AddReplayHandler(entityConfig, h); err != nil {
			return err
		}
		for _, commandConfig := range cfg.CommandConfigs {
			if err := GlobalRegistry.AddCommandHandler(commandConfig, h); err != nil {
				return err
			}
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
		h = UseCommandHandlerMiddleware(h, middlewares...)

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
	for _, handler := range handlers {
		commandHandles := NewCommandHandles(handler)
		h := NewCommandHandler(handler, commandHandles)
		h = UseCommandHandlerMiddleware(h, middlewares...)

		for t := range commandHandles {
			commandHandler := NewCommandConfig(t)
			if err := GlobalRegistry.AddCommandHandler(commandHandler, h); err != nil {
				return err
			}
		}
	}

	// sagas
	for _, saga := range sagas {
		handles := NewSagaHandles(saga)
		h := NewSagaEventHandler(handles, saga)

		for t := range handles {
			eventConfig := NewEventConfig(t)
			if err := GlobalRegistry.AddEventHandler(eventConfig, h); err != nil {
				return err
			}
		}
	}

	// projectors
	for _, projector := range projectors {
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
				eventConfig := NewEventConfig(evt)
				h := NewProjectorEventHandler(entityConfig, handles, projector)

				if err := GlobalRegistry.AddEventHandler(eventConfig, h); err != nil {
					return err
				}
			}
		}
	}

	// events
	for _, evt := range events {
		evtConfig := NewEventConfig(evt)
		if err := GlobalRegistry.AddEvent(evtConfig); err != nil {
			return err
		}
	}

	return nil
}
