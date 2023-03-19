package es

import (
	"fmt"
	"reflect"
	"strings"
)

type config struct {
	providerConfig  *ProviderConfig
	entities        map[string]*EntityConfig
	commands        map[string]*CommandConfig
	events          map[string]*EventConfig
	commandHandlers map[reflect.Type]CommandHandler
	eventHandlers   map[reflect.Type][]EventHandler
	replayHandlers  map[string]CommandHandler
}

func (c config) GetProviderConfig() *ProviderConfig {
	return c.providerConfig
}

func (c config) GetEntityConfigs() map[string]*EntityConfig {
	return c.entities
}
func (c config) GetCommandConfigs() map[string]*CommandConfig {
	return c.commands
}
func (c config) GetEventConfigs() map[string]*EventConfig {
	return c.events
}

func (c config) GetCommandHandlers() map[reflect.Type]CommandHandler {
	return c.commandHandlers
}
func (c config) GetEventHandlers() map[reflect.Type][]EventHandler {
	return c.eventHandlers
}

func (c config) GetReplayHandler(entityName string) CommandHandler {
	return c.replayHandlers[entityName]
}

type Builder interface {
	Build() (Config, error)

	SetProviderConfig(*ProviderConfig) Builder
	AddCommandHandler(IsCommandHandler) Builder
	AddSaga(IsSaga) Builder
	AddProjector(IsProjector) Builder
	AddAggregate(Aggregate) Builder
	AddEntity(Entity) Builder
	AddAggregateConfig(*AggregateConfig) Builder
	AddMiddleware(CommandHandlerMiddleware) Builder
}

// todo what about event handlers
// todo what about command handlers
type builder struct {
	providerConfig   *ProviderConfig
	handlers         []IsCommandHandler
	sagas            []IsSaga
	projectors       []IsProjector
	aggregates       []Aggregate
	entities         []Entity
	aggregateConfigs []*AggregateConfig
	middlewares      []CommandHandlerMiddleware
}

func (b *builder) SetProviderConfig(cfg *ProviderConfig) Builder {
	b.providerConfig = cfg
	return b
}

func (b *builder) AddCommandHandler(handler IsCommandHandler) Builder {
	b.handlers = append(b.handlers, handler)
	return b
}

func (b *builder) AddSaga(saga IsSaga) Builder {
	b.sagas = append(b.sagas, saga)
	return b
}

func (b *builder) AddProjector(p IsProjector) Builder {
	b.projectors = append(b.projectors, p)
	return b
}

func (b *builder) AddAggregate(agg Aggregate) Builder {
	b.aggregates = append(b.aggregates, agg)
	return b
}

func (b *builder) AddEntity(entity Entity) Builder {
	b.entities = append(b.entities, entity)
	return b
}

func (b *builder) AddAggregateConfig(cfg *AggregateConfig) Builder {
	b.aggregateConfigs = append(b.aggregateConfigs, cfg)
	return b
}

func (b *builder) AddMiddleware(m CommandHandlerMiddleware) Builder {
	b.middlewares = append(b.middlewares, m)
	return b
}

func (b *builder) Build() (Config, error) {
	entities := make(map[string]*EntityConfig)
	commands := make(map[string]*CommandConfig)
	events := make(map[string]*EventConfig)
	replayHandlers := make(map[string]CommandHandler)
	commandHandlers := make(map[reflect.Type]CommandHandler)
	eventHandlers := make(map[reflect.Type][]EventHandler)

	// register entities
	for _, entity := range b.entities {
		opts := NewEntityOptions(entity)
		entityConfig, err := NewEntityConfig(opts)
		if err != nil {
			return nil, err
		}
		name := strings.ToLower(entityConfig.Name)
		if _, ok := entities[name]; ok {
			return nil, fmt.Errorf("duplicate entity: %s", name)
		}
		entities[name] = entityConfig
	}

	// basic aggregates
	for _, cfg := range b.aggregateConfigs {
		entityConfig, err := NewEntityConfig(cfg.EntityOptions)
		if err != nil {
			return nil, err
		}
		name := strings.ToLower(entityConfig.Name)
		if _, ok := entities[name]; ok {
			return nil, fmt.Errorf("duplicate entity: %s", name)
		}
		entities[name] = entityConfig

		h := NewSourcedAggregateHandler(entityConfig, nil)
		h = UseCommandHandlerMiddleware(h, b.middlewares...)
		replayHandlers[name] = h

		for _, cmdCfg := range cfg.CommandConfigs {
			name := strings.ToLower(cmdCfg.Name)
			if _, ok := commands[name]; ok {
				return nil, fmt.Errorf("duplicate command: %s", name)
			}
			commands[name] = cmdCfg

			if _, ok := commandHandlers[cmdCfg.Type]; ok {
				return nil, fmt.Errorf("duplicate command handler: %s", cmdCfg.Type)
			}
			commandHandlers[cmdCfg.Type] = h
		}
	}

	// dynamic aggregates
	for _, agg := range b.aggregates {
		opts := NewEntityOptions(agg)
		entityConfig, err := NewEntityConfig(opts)
		if err != nil {
			return nil, err
		}
		name := strings.ToLower(entityConfig.Name)
		if _, ok := entities[name]; ok {
			return nil, fmt.Errorf("duplicate entity: %s", name)
		}
		entities[name] = entityConfig

		// handles!
		commandHandles := NewCommandHandles(agg)

		var commandConfigs []*CommandConfig
		for t := range commandHandles {
			cmdConfig := NewCommandConfig(t)
			commandConfigs = append(commandConfigs, cmdConfig)
		}

		h := NewSourcedAggregateHandler(entityConfig, commandHandles)
		h = UseCommandHandlerMiddleware(h, b.middlewares...)
		replayHandlers[name] = h

		for _, cmdCfg := range commandConfigs {
			name := strings.ToLower(cmdCfg.Name)
			if _, ok := commands[name]; ok {
				return nil, fmt.Errorf("duplicate command: %s", name)
			}
			commands[name] = cmdCfg

			if _, ok := commandHandlers[cmdCfg.Type]; ok {
				return nil, fmt.Errorf("duplicate command handler: %s", cmdCfg.Type)
			}
			commandHandlers[cmdCfg.Type] = h
		}
	}

	for _, handler := range b.handlers {
		// handles!
		handles := NewCommandHandles(handler)
		var commandConfigs []*CommandConfig
		for t := range handles {
			cmdConfig := NewCommandConfig(t)
			commandConfigs = append(commandConfigs, cmdConfig)
		}

		h := NewCommandHandler(handler, handles)
		h = UseCommandHandlerMiddleware(h, b.middlewares...)

		for _, cmdCfg := range commandConfigs {
			name := strings.ToLower(cmdCfg.Name)
			if _, ok := commands[name]; ok {
				return nil, fmt.Errorf("duplicate command: %s", name)
			}
			commands[name] = cmdCfg

			if _, ok := commandHandlers[cmdCfg.Type]; ok {
				return nil, fmt.Errorf("duplicate command handler: %s", cmdCfg.Type)
			}
			commandHandlers[cmdCfg.Type] = h
		}
	}

	for _, saga := range b.sagas {
		handles := NewSagaHandles(saga)

		eventConfigs := []*EventConfig{}
		for t := range handles {
			evtConfig := NewEventConfig(t)
			eventConfigs = append(eventConfigs, evtConfig)
		}

		h := NewSagaEventHandler(handles, saga)

		// only add our event configs once
		for _, evtCfg := range eventConfigs {
			name := strings.ToLower(evtCfg.Name)
			if _, ok := events[name]; !ok {
				events[name] = evtCfg
			}
			eventHandlers[evtCfg.Type] = append(eventHandlers[evtCfg.Type], h)
		}
	}

	for _, projector := range b.projectors {
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
				h := NewProjectorEventHandler(entityConfig, handles, projector)

				evtCfg := NewEventConfig(evt)
				name := strings.ToLower(evtCfg.Name)
				if _, ok := events[name]; !ok {
					events[name] = evtCfg
				}
				eventHandlers[evtCfg.Type] = append(eventHandlers[evtCfg.Type], h)
			}
		}
	}

	return &config{
		providerConfig:  b.providerConfig,
		entities:        entities,
		commands:        commands,
		events:          events,
		commandHandlers: commandHandlers,
		eventHandlers:   eventHandlers,
		replayHandlers:  replayHandlers,
	}, nil
}

func NewBuilder() Builder {
	b := &builder{}
	return b
}
