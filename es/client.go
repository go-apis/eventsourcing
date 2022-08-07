package es

import (
	"context"
	"fmt"
	"reflect"

	"github.com/contextcloud/eventstore/es/utils"
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

	entities        []*Entity
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
	eventHandlers := c.cfg.GetEventHandlers()
	for _, eh := range eventHandlers {
		h := eh.GetHandler()
		for _, evt := range eh.GetEvents() {
			t := utils.GetElemType(evt)
			c.eventHandlers[t] = append(c.eventHandlers[t], h)
		}
	}

	commandHandlers := c.cfg.GetCommandHandlers()
	for _, ch := range commandHandlers {
		// ent := ch.Factory()
		// t := utils.GetElemType(ent)

		// c.entities = append(c.entities, ent)
		// c.entityOptions[t] = &agg.EntityOptions

		opts := ch.GetEntityOptions()
		c.entityOptions[opts.Name] = &opts

		handler := ch.GetHandler()
		for _, cmd := range ch.GetCommands() {
			t := utils.GetElemType(cmd)
			c.commandHandlers[t] = handler
		}
	}

	return nil
}

func (c *client) HandleCommands(ctx context.Context, cmds ...Command) error {
	for _, cmd := range cmds {
		t := utils.GetElemType(cmd)
		h, ok := c.commandHandlers[t]
		if !ok {
			return fmt.Errorf("command handler not found: %v", t)
		}
		if err := h.Handle(ctx, cmd); err != nil {
			return err
		}
	}
	return nil
}

func (c *client) HandleEvents(ctx context.Context, evts ...Event) error {
	for _, evt := range evts {
		t := utils.GetElemType(evt.Data)
		all, ok := c.eventHandlers[t]
		if !ok {
			continue
		}

		for _, h := range all {
			if err := h.Handle(ctx, evt, evt.Data); err != nil {
				return err
			}
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
