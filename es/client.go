package es

import (
	"context"
	"strings"
)

type Client interface {
	Unit(ctx context.Context) (Unit, error)
}

type client struct {
	providerConfig *ProviderConfig
	conn           Conn
	streamer       Streamer
}

func (c *client) Unit(ctx context.Context) (Unit, error) {
	// if we already have a unit, return it
	if unit, err := GetUnit(ctx); err == nil {
		return unit, nil
	}

	// create it.
	unit, err := newUnit(ctx, c)
	if err != nil {
		return nil, err
	}
	return unit, nil
}

func NewClient(ctx context.Context, pcfg *ProviderConfig) (Client, error) {
	conn, err := GetConn(ctx, pcfg)
	if err != nil {
		return nil, err
	}

	streamer, err := GetStreamer(ctx, pcfg)
	if err != nil {
		return nil, err
	}

	cli := &client{
		providerConfig: pcfg,
		conn:           conn,
		streamer:       streamer,
	}

	callback := func(ctx context.Context, evt *Event) error {
		// we don't hand events from ourself
		if strings.EqualFold(evt.Service, pcfg.Service) {
			return nil
		}

		// create the unit.
		unit, err := cli.Unit(ctx)
		if err != nil {
			return err
		}
		return unit.Handle(ctx, evt)
	}

	if err := streamer.Start(ctx, callback); err != nil {
		return nil, err
	}
	return cli, nil
}
