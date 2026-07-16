package es

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

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
	MigrateDb(ctx context.Context) error
}

type client struct {
	providerConfig *ProviderConfig
	registry       Registry
	conn           Conn
	publisher      EventPublisher
	relay          *outboxRelay
}

func (c *client) Unit(ctx context.Context) (Unit, error) {
	// if we already have a unit, return it
	if unit, err := GetUnit(ctx); err == nil {
		return unit, nil
	}

	// create it.
	unit, err := newUnit(ctx, c.providerConfig.Service, c.registry, c.conn, c.relay)
	if err != nil {
		return nil, err
	}
	return unit, nil
}

func (c *client) MigrateDb(ctx context.Context) error {
	return c.conn.MigrateDb(ctx)
}

// PushHandler returns the streamer's HTTP delivery handler when the
// configured streamer consumes via push (see PushReceiver).
func (c *client) PushHandler() (http.Handler, bool) {
	if pr, ok := c.publisher.(PushReceiver); ok {
		return pr.PushHandler(), true
	}
	return nil, false
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

	var scheduler CommandScheduler
	if pcfg.UseScheduler {
		scheduler, err = NewCommandScheduler(ctx, client)
		if err != nil {
			return nil, err
		}
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
				// The shared topic carries every service's published events;
				// ones this service has no registration for can never be
				// handled, so skip them instead of nacking into a redelivery
				// loop that jams the subscription.
				if errors.Is(err, ErrNotFound) {
					return nil
				}
				return err
			}

			innerCtx := ctx
			if evt.By != nil {
				innerCtx = SetActor(ctx, evt.By)
			}
			// Handlers can mint detached units (chunked commits) from ctx.
			innerCtx = SetClient(innerCtx, client)

			// create the unit.
			unit, err := client.Unit(innerCtx)
			if err != nil {
				return err
			}
			if err := unit.Handle(innerCtx, group, evt); err != nil {
				slog.ErrorContext(innerCtx, "bus delivery failed",
					"group", group,
					"event", evt.Type,
					"aggregate_id", evt.AggregateId,
					"error", err,
				)
				return err
			}
			return nil
		})
		if err := streamer.AddHandler(ctx, name, handler); err != nil {
			return nil, err
		}
	}

	client.publisher = streamer

	// The relay owns delivery to the stream: units park publishable events
	// in the outbox and nudge it. It also drains rows a crashed process
	// left behind.
	relay, err := newOutboxRelay(ctx, pcfg.Service, conn, reg, streamer)
	if err != nil {
		return nil, err
	}
	client.relay = relay

	// close stuff if we have an error.
	defer func() {
		if err != nil {
			if relay != nil {
				_ = relay.Close(ctx)
			}
			if streamer != nil {
				_ = streamer.Close(ctx)
			}
			if conn != nil {
				_ = conn.Close(ctx)
			}
			if scheduler != nil {
				_ = scheduler.Close(ctx)
			}
		}
	}()

	go relay.run(ctx)

	go func() {
		<-ctx.Done()

		ctx := context.Background()
		if relay != nil {
			_ = relay.Close(ctx)
		}
		if streamer != nil {
			_ = streamer.Close(ctx)
		}
		if conn != nil {
			_ = conn.Close(ctx)
		}
		if scheduler != nil {
			_ = scheduler.Close(ctx)
		}
	}()

	return client, nil
}
