package es

import (
	"context"
	"fmt"
	"reflect"

	"github.com/neilotoole/errgroup"
)

var ErrNotEventHandler = fmt.Errorf("not a event handler")

type Publisher interface {
	PublishAsync(ctx context.Context, evts ...Event) error
}

type publisher struct {
	eventHandlers map[reflect.Type][]EventHandler
}

func (c *publisher) PublishAsync(ctx context.Context, evts ...Event) error {
	numG, qSize := 8, 4
	g, ctx := errgroup.WithContextN(ctx, numG, qSize)

	for _, evt := range evts {
		t := reflect.TypeOf(evt.Data)
		handlers := c.eventHandlers[t]

		for _, h := range handlers {
			d := evt
			in := h
			g.Go(func() error {
				return in.Handle(ctx, d, d.Data)
			})
		}
	}

	return g.Wait()
}
