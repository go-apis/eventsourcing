package es

import (
	"context"

	"github.com/google/uuid"
)

type Client interface {
	Unit(ctx context.Context) (Unit, error)
}

type clientEventHandler struct {
	client Client
	group  string
}

func (h *clientEventHandler) Handle(ctx context.Context, evt *Event) error {
	// create the unit.
	unit, err := h.client.Unit(ctx)
	if err != nil {
		return err
	}
	return unit.Handle(ctx, h.group, evt)
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

func NewClient(ctx context.Context, pcfg *ProviderConfig) (cli Client, err error) {
	streamer, err := GetStreamer(ctx, pcfg)
	if err != nil {
		return nil, err
	}

	conn, err := GetConn(ctx, pcfg)
	if err != nil {
		return nil, err
	}

	cli = &client{
		providerConfig: pcfg,
		conn:           conn,
		streamer:       streamer,
	}

	// close stuff if we have an error.
	defer func() {
		if err != nil {
			streamer.Close(ctx)
		}
	}()

	// get the groups.
	groups := GlobalRegistry.GetEventHandlerGroups()
	for _, group := range groups {
		if group == InternalGroup {
			continue
		}

		name := ""
		if group == RandomGroup {
			name = uuid.NewString()
		}

		handler := &clientEventHandler{
			client: cli,
			group:  group,
		}
		if err := streamer.AddHandler(ctx, name, handler); err != nil {
			return nil, err
		}
	}
	return
}
