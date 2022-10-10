package es

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"go.opentelemetry.io/otel"
)

type Client interface {
	GetServiceName() string
	GetEntityConfig(name string) (*EntityConfig, error)
	GetCommandConfig(name string) (*CommandConfig, error)
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
	commandConfigs  map[string]*CommandConfig
	commandHandlers map[reflect.Type]CommandHandler
	eventHandlers   map[reflect.Type][]EventHandler
}

func (c *client) GetServiceName() string {
	pcfg := c.cfg.GetProviderConfig()
	return pcfg.ServiceName
}

func (c *client) GetEntityConfig(name string) (*EntityConfig, error) {
	if cfg, ok := c.entityConfigs[strings.ToLower(name)]; ok {
		return cfg, nil
	}

	return nil, fmt.Errorf("entity config not found: %s", name)
}

func (c *client) GetCommandConfig(name string) (*CommandConfig, error) {
	if cfg, ok := c.commandConfigs[strings.ToLower(name)]; ok {
		return cfg, nil
	}

	return nil, fmt.Errorf("command config not found: %s", name)
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

func (c *client) initialize(ctx context.Context) error {
	pctx, pspan := otel.Tracer("client").Start(ctx, "Initialize")
	defer pspan.End()

	pcfg := c.cfg.GetProviderConfig()

	eventHandlers := c.cfg.GetEventHandlers()
	for _, eh := range eventHandlers {
		handler := eh.handler
		for _, evt := range eh.eventConfigs {
			c.eventHandlers[evt.Type] = append(c.eventHandlers[evt.Type], handler)
		}
	}

	commandHandlers := c.cfg.GetCommandHandlers()
	commandHandlerMiddlewares := c.cfg.GetCommandHandlerMiddlewares()
	for _, ch := range commandHandlers {
		handler := UseCommandHandlerMiddleware(ch.handler, commandHandlerMiddlewares...)
		for _, cmd := range ch.commandConfigs {
			c.commandHandlers[cmd.Type] = handler
		}
	}

	entityConfigs := c.cfg.GetEntityConfigs()
	for _, entityConfig := range entityConfigs {
		key := strings.ToLower(entityConfig.Name)
		c.entityConfigs[key] = entityConfig
	}
	commandConfigs := c.cfg.GetCommandConfigs()
	for _, commandConfig := range commandConfigs {
		key := strings.ToLower(commandConfig.Name)
		c.commandConfigs[key] = commandConfig
	}
	eventConfigs := c.cfg.GetEventConfigs()

	initOpts := InitializeOptions{
		ServiceName:   pcfg.ServiceName,
		EntityConfigs: entityConfigs,
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
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

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

func NewClient(ctx context.Context, cfg Config) (Client, error) {
	pcfg := cfg.GetProviderConfig()
	if pcfg == nil {
		return nil, fmt.Errorf("provider config not set")
	}

	conn, err := GetConn(pcfg)
	if err != nil {
		return nil, err
	}

	streamer, err := GetStreamer(pcfg)
	if err != nil {
		return nil, err
	}

	cli := &client{
		cfg:             cfg,
		conn:            conn,
		streamer:        streamer,
		entityConfigs:   make(map[string]*EntityConfig),
		commandConfigs:  make(map[string]*CommandConfig),
		commandHandlers: map[reflect.Type]CommandHandler{},
		eventHandlers:   map[reflect.Type][]EventHandler{},
	}

	if err := cli.initialize(ctx); err != nil {
		return nil, err
	}

	return cli, nil
}
