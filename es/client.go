package es

import (
	"context"

	"github.com/google/uuid"
)

func GenerateName(group string) string {
	switch group {
	case InternalGroup:
		return ""
	case ExternalGroup:
		return ""
	case RandomGroup:
		return uuid.NewString()
	default:
		return group
	}
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

	scheduler, err := NewCommandScheduler(ctx, client)
	if err != nil {
		return nil, err
	}

	streamer, err := GetStreamer(ctx, pcfg)
	if err != nil {
		return nil, err
	}

	for _, group := range reg.GetGroups() {
		if group == InternalGroup {
			continue
		}

		name := GenerateName(group)
		handler := MessageHandler(func(ctx context.Context, payload []byte) error {
			evt, err := reg.ParseEvent(ctx, payload)
			if err != nil {
				return err
			}

			innerCtx := ctx
			if evt.By != nil {
				innerCtx = SetActor(ctx, evt.By)
			}

			// create the unit.
			unit, err := client.Unit(innerCtx)
			if err != nil {
				return err
			}
			return unit.Handle(innerCtx, group, evt)
		})
		if err := streamer.AddHandler(ctx, name, handler); err != nil {
			return nil, err
		}
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
