package es

import (
	"context"
	"fmt"
	"reflect"

	"github.com/contextcloud/eventstore/es/utils"
	"go.opentelemetry.io/otel"
)

type Client interface {
	GetServiceName() string
	GetEntityOptions(name string) (*EntityOptions, error)
	Initialize(ctx context.Context) error
	Unit(ctx context.Context) (Unit, error)

	HandleCommands(ctx context.Context, cmds ...Command) error
	HandleEvents(ctx context.Context, events ...Event) error
	PublishEvents(ctx context.Context, events ...Event) error
}

type client struct {
	cfg  Config
	conn Conn

	entityOptions   map[string]*EntityOptions
	commandHandlers map[reflect.Type]CommandHandler
	eventHandlers   map[reflect.Type][]EventHandler
}

func (c *client) GetServiceName() string {
	return c.cfg.GetServiceName()
}

func (c *client) GetEntityOptions(name string) (*EntityOptions, error) {
	if opts, ok := c.entityOptions[name]; ok {
		return opts, nil
	}

	return nil, fmt.Errorf("entity options not found: %s", name)
}

func (c *client) Unit(ctx context.Context) (Unit, error) {
	if unit := UnitFromContext(ctx); unit != nil {
		return unit, nil
	}

	data, err := c.conn.NewData(ctx)
	if err != nil {
		return nil, err
	}

	return newUnit(c, data)
}

func (c *client) Initialize(ctx context.Context) error {
	serviceName := c.cfg.GetServiceName()

	events := make(map[reflect.Type]bool)
	eventHandlers := c.cfg.GetEventHandlers()
	for _, eh := range eventHandlers {
		handler := eh.GetHandler()
		for _, evt := range eh.GetEvents() {
			events[evt] = true
			c.eventHandlers[evt] = append(c.eventHandlers[evt], handler)
		}
	}

	var allEntityOpts []EntityOptions
	entities := c.cfg.GetEntities()
	for _, e := range entities {
		opts := e.GetOptions()
		allEntityOpts = append(allEntityOpts, opts)
		c.entityOptions[opts.Name] = &opts
	}

	commandHandlers := c.cfg.GetCommandHandlers()
	for _, ch := range commandHandlers {
		handler := ch.GetHandler()
		for _, cmd := range ch.GetCommands() {
			c.commandHandlers[cmd] = handler
		}
	}

	var allEventOpts []EventOptions
	for t := range events {
		name := t.String()
		allEventOpts = append(allEventOpts, EventOptions{
			Name: name,
			Type: t,
		})
	}

	initOpts := InitializeOptions{
		ServiceName:   serviceName,
		EntityOptions: allEntityOpts,
		EventOptions:  allEventOpts,
	}

	stream, err := c.conn.Initialize(ctx, initOpts)
	if err != nil {
		return err
	}
	go c.handleStream(ctx, stream)

	return nil
}

func (c *client) handleStream(ctx context.Context, stream *Stream) {
	// TODO: handle stream

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

func (c *client) handleEvent(ctx context.Context, evt Event) error {
	pctx, pspan := otel.Tracer("client").Start(ctx, "HandleEvent")
	defer pspan.End()

	t := utils.GetElemType(evt.Data)
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
func (c *client) eventHandlerHandleEvent(ctx context.Context, h EventHandler, evt Event) error {
	pctx, pspan := otel.Tracer("client").Start(ctx, "EventHandlerHandleEvent")
	defer pspan.End()

	if err := h.Handle(pctx, evt); err != nil {
		return err
	}
	return nil
}

func (c *client) HandleEvents(ctx context.Context, evts ...Event) error {
	pctx, pspan := otel.Tracer("client").Start(ctx, "HandleEvents")
	defer pspan.End()

	for _, evt := range evts {
		if err := c.handleEvent(pctx, evt); err != nil {
			return err
		}
	}
	return nil
}

func (c *client) PublishEvents(ctx context.Context, evts ...Event) error {
	return c.conn.Publish(ctx, evts...)
}

func NewClient(cfg Config, conn Conn) (Client, error) {
	cli := &client{
		cfg:             cfg,
		conn:            conn,
		entityOptions:   map[string]*EntityOptions{},
		commandHandlers: map[reflect.Type]CommandHandler{},
		eventHandlers:   map[reflect.Type][]EventHandler{},
	}

	ctx := context.Background()
	if err := cli.Initialize(ctx); err != nil {
		return nil, err
	}

	return cli, nil
}
