package es

import (
	"context"
	"fmt"
	"reflect"

	"go.opentelemetry.io/otel"
)

type Client interface {
	GetServiceName() string
	GetEntityConfig(name string) (*EntityConfig, error)
	Initialize(ctx context.Context) error
	Unit(ctx context.Context) (Unit, error)

	HandleCommands(ctx context.Context, cmds ...Command) error
	HandleEvents(ctx context.Context, events ...*Event) error
	PublishEvents(ctx context.Context, events ...*Event) error
}

type client struct {
	cfg      Config
	conn     Conn
	streamer Streamer

	entityConfigs   map[string]*EntityConfig
	commandHandlers map[reflect.Type]CommandHandler
	eventHandlers   map[reflect.Type][]EventHandler
}

func (c *client) GetServiceName() string {
	return c.cfg.GetServiceName()
}

func (c *client) GetEntityConfig(name string) (*EntityConfig, error) {
	if opts, ok := c.entityConfigs[name]; ok {
		return opts, nil
	}

	return nil, fmt.Errorf("entity config not found: %s", name)
}

func (c *client) Unit(ctx context.Context) (Unit, error) {
	pctx, pspan := otel.Tracer("client").Start(ctx, "Unit")
	defer pspan.End()

	if unit, err := GetUnit(pctx); err == nil {
		return unit, nil
	}

	data, err := c.conn.NewData(pctx)
	if err != nil {
		return nil, err
	}

	return newUnit(c, data)
}

func (c *client) Initialize(ctx context.Context) error {
	pctx, pspan := otel.Tracer("client").Start(ctx, "Initialize")
	defer pspan.End()

	serviceName := c.cfg.GetServiceName()

	events := make(map[reflect.Type]bool)
	eventHandlers := c.cfg.GetEventHandlers()
	for _, eh := range eventHandlers {
		handler := eh.handler
		for _, evt := range eh.events {
			events[evt] = true
			c.eventHandlers[evt] = append(c.eventHandlers[evt], handler)
		}
	}

	entities := c.cfg.GetEntities()
	for _, entityConfig := range entities {
		c.entityConfigs[entityConfig.Name] = entityConfig
	}

	commandHandlers := c.cfg.GetCommandHandlers()
	for _, ch := range commandHandlers {
		handler := ch.handler
		for _, cmd := range ch.commands {
			c.commandHandlers[cmd] = handler
		}
	}

	eventConfigs := []*EventConfig{}
	for t := range events {
		r := t
		for r.Kind() == reflect.Ptr {
			r = r.Elem()
		}
		name := r.Name()
		factory := func() (interface{}, error) {
			out := reflect.New(r).Interface()
			return out, nil
		}
		eventConfigs = append(eventConfigs, &EventConfig{
			Name:    name,
			Type:    r,
			Factory: factory,
		})
	}

	initOpts := InitializeOptions{
		ServiceName:   serviceName,
		EntityConfigs: entities,
		EventConfigs:  eventConfigs,
	}

	if err := c.conn.Initialize(pctx, initOpts); err != nil {
		return err
	}

	if c.streamer != nil {
		if err := c.streamer.Start(pctx, initOpts, c.handleStreamEvent); err != nil {
			return err
		}
	}

	return nil
}
func (c *client) handleStreamEvent(ctx context.Context, evt *Event) error {
	pctx, pspan := otel.Tracer("client").Start(ctx, "HandleStreamEvent")
	defer pspan.End()

	unit, err := c.Unit(pctx)
	if err != nil {
		return err
	}
	pctx = SetUnit(pctx, unit)
	pctx = SetNamespace(pctx, evt.Namespace)

	// create the transaction!
	tx, err := unit.NewTx(pctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(pctx)

	if err := c.handleEvent(pctx, evt); err != nil {
		return err
	}

	if _, err := tx.Commit(pctx); err != nil {
		return err
	}

	return nil
}

func (c *client) handleCommand(ctx context.Context, cmd Command) error {
	pctx, pspan := otel.Tracer("client").Start(ctx, "HandleCommand")
	defer pspan.End()

	t := reflect.TypeOf(cmd)
	h, ok := c.commandHandlers[t]
	if !ok {
		return fmt.Errorf("command handler not found: %v", t)
	}
	if err := h.Handle(pctx, cmd); err != nil {
		return err
	}
	return nil
}

func (c *client) HandleCommands(ctx context.Context, cmds ...Command) error {
	pctx, pspan := otel.Tracer("client").Start(ctx, "HandleCommands")
	defer pspan.End()

	for _, cmd := range cmds {
		if err := c.handleCommand(pctx, cmd); err != nil {
			return err
		}
	}
	return nil
}

func (c *client) handleEvent(ctx context.Context, evt *Event) error {
	pctx, pspan := otel.Tracer("client").Start(ctx, "HandleEvent")
	defer pspan.End()

	t := reflect.TypeOf(evt.Data)
	all, ok := c.eventHandlers[t]
	if !ok {
		return nil
	}

	for _, h := range all {
		if err := c.eventHandlerHandleEvent(pctx, h, evt); err != nil {
			return err
		}
	}

	return nil
}
func (c *client) eventHandlerHandleEvent(ctx context.Context, h EventHandler, evt *Event) error {
	pctx, pspan := otel.Tracer("client").Start(ctx, "EventHandlerHandleEvent")
	defer pspan.End()

	if err := h.Handle(pctx, evt); err != nil {
		return err
	}
	return nil
}

func (c *client) HandleEvents(ctx context.Context, evts ...*Event) error {
	pctx, pspan := otel.Tracer("client").Start(ctx, "HandleEvents")
	defer pspan.End()

	for _, evt := range evts {
		if err := c.handleEvent(pctx, evt); err != nil {
			return err
		}
	}
	return nil
}

func (c *client) PublishEvents(ctx context.Context, evts ...*Event) error {
	pctx, pspan := otel.Tracer("client").Start(ctx, "PublishEvents")
	defer pspan.End()

	if c.streamer == nil {
		return nil
	}

	return c.streamer.Publish(pctx, evts...)
}

func NewClient(cfg Config, conn Conn, streamer Streamer) (Client, error) {
	cli := &client{
		cfg:             cfg,
		conn:            conn,
		streamer:        streamer,
		entityConfigs:   make(map[string]*EntityConfig),
		commandHandlers: map[reflect.Type]CommandHandler{},
		eventHandlers:   map[reflect.Type][]EventHandler{},
	}

	return cli, nil
}
