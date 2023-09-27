package es

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type Client interface {
	GetService() string
	GetEntityConfig(name string) (*EntityConfig, error)
	GetCommandConfig(name string) (*CommandConfig, error)
	Unit(ctx context.Context) (Unit, error)

	ReplayCommands(ctx context.Context, cmds ...*ReplayCommand) error
	HandleCommands(ctx context.Context, cmds ...Command) error
	HandleEvents(ctx context.Context, events ...*Event) error
	PublishEvents(ctx context.Context, events ...*Event) error
}

type client struct {
	cfg      Config
	conn     Conn
	streamer Streamer
}

func (c *client) GetService() string {
	pcfg := c.cfg.GetProviderConfig()
	return pcfg.Service
}

func (c *client) GetEntityConfig(name string) (*EntityConfig, error) {
	if cfg, ok := c.cfg.GetEntityConfigs()[strings.ToLower(name)]; ok {
		return cfg, nil
	}

	return nil, fmt.Errorf("entity config not found: %s", name)
}

func (c *client) GetCommandConfig(name string) (*CommandConfig, error) {
	if cfg, ok := c.cfg.GetCommandConfigs()[strings.ToLower(name)]; ok {
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

	return newUnit(pctx, c, data)
}

func (c *client) initialize(ctx context.Context) error {
	if err := c.conn.Initialize(ctx, c.cfg); err != nil {
		return err
	}

	if c.streamer != nil {
		if err := c.streamer.Start(ctx, c.cfg, c.handleStreamEvent); err != nil {
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

	defer unit.Rollback(pctx)

	if err := c.handleEvent(pctx, evt); err != nil {
		return err
	}

	if _, err := unit.Commit(pctx); err != nil {
		return err
	}

	return nil
}

func (c *client) replayCommand(ctx context.Context, cmd *ReplayCommand) error {
	pctx, pspan := otel.Tracer("client").Start(ctx, "ReplayCommand")
	defer pspan.End()

	ns := cmd.GetNamespace()
	if ns != "" {
		pctx = SetNamespace(pctx, ns)
	}

	pspan.SetAttributes(
		attribute.String("name ", cmd.AggregateName),
		attribute.String("id", cmd.GetAggregateId().String()),
	)

	h := c.cfg.GetReplayHandler(cmd.AggregateName)
	if h == nil {
		return fmt.Errorf("aggregate command handler not found: %v", cmd.AggregateName)
	}

	if err := h.Handle(pctx, cmd); err != nil {
		return err
	}
	return nil
}
func (c *client) ReplayCommands(ctx context.Context, cmds ...*ReplayCommand) error {
	for _, cmd := range cmds {
		if err := c.replayCommand(ctx, cmd); err != nil {
			return err
		}
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

	h, ok := c.cfg.GetCommandHandlers()[t]
	if !ok {
		return fmt.Errorf("command handler not found: %v", t)
	}

	// check if we have a namespace command
	if nsCmd, ok := cmd.(NamespaceCommand); ok {
		ns := nsCmd.GetNamespace()
		if ns != "" {
			pctx = SetNamespace(pctx, ns)
		}
	}

	pspan.SetAttributes(
		attribute.String("command", t.Name()),
		attribute.String("id", cmd.GetAggregateId().String()),
	)

	if err := h.Handle(pctx, cmd); err != nil {
		return err
	}
	return nil
}
func (c *client) HandleCommands(ctx context.Context, cmds ...Command) error {
	for _, cmd := range cmds {
		if err := c.handleCommand(ctx, cmd); err != nil {
			return err
		}
	}
	return nil
}

func (c *client) handleEvent(ctx context.Context, evt *Event) error {
	t := reflect.TypeOf(evt.Data)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	all := c.cfg.GetEventHandlers()[t]
	for _, h := range all {
		if err := c.eventHandlerHandleEvent(ctx, h, evt); err != nil {
			return err
		}
	}

	return nil
}
func (c *client) eventHandlerHandleEvent(ctx context.Context, h EventHandler, evt *Event) error {
	pctx, pspan := otel.Tracer("client").Start(ctx, "EventHandlerHandleEvent")
	defer pspan.End()

	pspan.SetAttributes(
		attribute.String("event", evt.Type),
		attribute.String("id", evt.AggregateId.String()),
		attribute.String("type", evt.AggregateType),
	)

	if err := h.Handle(pctx, evt); err != nil {
		return err
	}
	return nil
}
func (c *client) HandleEvents(ctx context.Context, evts ...*Event) error {
	for _, evt := range evts {
		if err := c.handleEvent(ctx, evt); err != nil {
			return err
		}
	}
	return nil
}
func (c *client) PublishEvents(ctx context.Context, evts ...*Event) error {
	if c.streamer == nil {
		return nil
	}

	configs := c.cfg.GetEventConfigs()

	var publishEvts []*Event
	for _, evt := range evts {
		cfg, ok := configs[strings.ToLower(evt.Type)]
		if !ok {
			continue
		}

		if cfg.Publish {
			publishEvts = append(publishEvts, evt)
		}
	}

	return c.streamer.Publish(ctx, publishEvts...)
}

func NewClient(ctx context.Context, cfg Config) (Client, error) {
	pcfg := cfg.GetProviderConfig()
	if pcfg == nil {
		return nil, fmt.Errorf("provider config not set")
	}

	conn, err := GetConn(ctx, pcfg)
	if err != nil {
		return nil, err
	}

	streamer, err := GetStreamer(ctx, pcfg)
	if err != nil {
		return nil, err
	}

	cli := &client{
		cfg:      cfg,
		conn:     conn,
		streamer: streamer,
	}

	if err := cli.initialize(ctx); err != nil {
		return nil, err
	}

	return cli, nil
}
