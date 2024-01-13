package es

import (
	"context"
)

type clientGroupMessageHandler struct {
	cli      Client
	registry Registry
}

func (c *clientGroupMessageHandler) HandleGroupMessage(ctx context.Context, group string, msg []byte) error {
	evt, err := c.registry.ParseEvent(ctx, msg)
	if err != nil {
		return err
	}

	// create the unit.
	unit, err := c.cli.Unit(ctx)
	if err != nil {
		return err
	}
	return unit.Handle(ctx, group, evt)
}

type clientCommandHandler struct {
	cli Client
}

func (c *clientCommandHandler) HandleCommand(ctx context.Context, cmd Command) error {
	// create the unit.
	unit, err := c.cli.Unit(ctx)
	if err != nil {
		return err
	}
	return unit.Dispatch(ctx, cmd)
}

type Client interface {
	Unit(ctx context.Context) (Unit, error)
}

type client struct {
	providerConfig *ProviderConfig
	registry       Registry
	conn           Conn
	publisher      EventPublisher
}

func (c *client) Unit(ctx context.Context) (Unit, error) {
	// if we already have a unit, return it
	if unit, err := GetUnit(ctx); err == nil {
		return unit, nil
	}

	// create it.
	unit, err := newUnit(ctx, c.providerConfig.Service, c.registry, c.conn, c.publisher)
	if err != nil {
		return nil, err
	}
	return unit, nil
}

func NewClient(ctx context.Context, pcfg *ProviderConfig, reg Registry) (cli Client, err error) {
	conn, err := GetConn(ctx, pcfg, reg)
	if err != nil {
		return nil, err
	}

	client := &client{
		providerConfig: pcfg,
		registry:       reg,
		conn:           conn,
	}

	groupMessageHandler := &clientGroupMessageHandler{
		cli:      client,
		registry: reg,
	}
	commandHandler := &clientCommandHandler{
		cli: client,
	}

	scheduler, err := NewCommandScheduler(ctx, reg, conn, commandHandler)
	if err != nil {
		return nil, err
	}

	streamer, err := GetStreamer(ctx, pcfg, reg, groupMessageHandler)
	if err != nil {
		return nil, err
	}

	client.publisher = streamer

	// close stuff if we have an error.
	defer func() {
		if err != nil {
			if streamer != nil {
				streamer.Close(ctx)
			}
			if conn != nil {
				conn.Close(ctx)
			}
			if scheduler != nil {
				scheduler.Close(ctx)
			}
		}
	}()
	go func() {
		<-ctx.Done()

		ctx := context.Background()
		if streamer != nil {
			streamer.Close(ctx)
		}
		if conn != nil {
			conn.Close(ctx)
		}
		if scheduler != nil {
			scheduler.Close(ctx)
		}
	}()

	return client, nil
}
