package es

import (
	"fmt"
	"reflect"
)

type Registry interface {
	EntityRegistry
	CommandRegistry
	EventRegistry
}

type registry struct {
	EntityRegistry
	CommandRegistry
	EventRegistry
}

func NewRegistry(service string, items ...interface{}) (Registry, error) {
	var sagas []IsSaga
	var projectors []IsProjector
	var eventHandlers []IsEventHandler
	var commandHandlers []IsCommandHandler
	var aggregates []Aggregate
	var entities []Entity
	var events []interface{}
	var middlewares []CommandHandlerMiddleware

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
		case CommandHandlerMiddleware:
			middlewares = append(middlewares, raw)
			continue
		default:
			return nil, fmt.Errorf("invalid item type: %T", item)
		}
	}

	entityRegistry := NewEntityRegistry()
	commandRegistry := NewCommandRegistry()
	eventRegistry := NewEventRegistry()

	// register entities
	for _, entity := range entities {
		opts := NewEntityOptions(entity)
		entityConfig, err := NewEntityConfig(opts)
		if err != nil {
			return nil, err
		}
		if err := entityRegistry.AddEntity(entityConfig); err != nil {
			return nil, err
		}
	}

	// events
	for _, evt := range events {
		evtConfig := NewEventConfig(service, evt)
		if err := eventRegistry.AddEvent(evtConfig); err != nil {
			return nil, err
		}
	}

	// dynamic aggregates
	for _, agg := range aggregates {
		opts := NewEntityOptions(agg)
		entityConfig, err := NewEntityConfig(opts)
		if err != nil {
			return nil, err
		}
		// handles!
		commandHandles := NewCommandHandles(agg)
		h := NewAggregateHandler(entityConfig, commandHandles)

		if err := entityRegistry.AddEntity(entityConfig); err != nil {
			return nil, err
		}

		if err := commandRegistry.SetReplayHandler(h, entityConfig); err != nil {
			return nil, err
		}
		for t := range commandHandles {
			commandConfig := NewCommandConfig(t)
			if err := commandRegistry.SetCommandHandler(h, commandConfig); err != nil {
				return nil, err
			}
		}
	}

	// handlers
	for _, commandHandler := range commandHandlers {
		commandHandles := NewCommandHandles(commandHandler)
		h := NewCommandHandler(commandHandler, commandHandles)

		for t := range commandHandles {
			commandConfig := NewCommandConfig(t)
			if err := commandRegistry.SetCommandHandler(h, commandConfig); err != nil {
				return nil, err
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
			if err := eventRegistry.AddGroupEventHandler(h, eventHandlerConfig.Group, eventConfig); err != nil {
				return nil, err
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
				return nil, err
			}
			for evt, handles := range m {
				eventConfig := NewEventConfig(service, evt)
				h := NewProjectorEventHandler(entityConfig, handles, projector)

				if err := eventRegistry.AddGroupEventHandler(h, eventHandlerConfig.Group, eventConfig); err != nil {
					return nil, err
				}
			}
		}
	}

	// eventhandlers
	for _, eventHandler := range eventHandlers {
		eventHandlerConfig := NewEventHandlerConfig(eventHandler)
		handles := NewEventHandlerHandles(eventHandler)
		h := NewEventHandler(eventHandler, handles)

		for t := range handles {
			eventConfig := NewEventConfig(service, t)
			if err := eventRegistry.AddGroupEventHandler(h, eventHandlerConfig.Group, eventConfig); err != nil {
				return nil, err
			}
		}
	}

	reg := &registry{
		EntityRegistry:  entityRegistry,
		CommandRegistry: commandRegistry,
		EventRegistry:   eventRegistry,
	}
	return reg, nil
}
