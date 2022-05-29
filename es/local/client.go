package local

import (
	"context"
	"eventstore/es"
)

type client struct {
	data        es.Data
	serviceName string
}

func (c *client) GetTx(ctx context.Context) (es.Tx, error) {
	return c.data.GetTx(ctx)
}
func (c *client) WithTx(ctx context.Context) (context.Context, es.Tx, error) {
	return c.data.WithTx(ctx)
}

func (c *client) NewSourcedStore(dispatcher es.Dispatcher, name string) es.SourcedStore {
	return newSourcedStore(c.data, c.serviceName, dispatcher, name)
}

func NewClient(data es.Data, serviceName string) es.Client {
	return &client{
		data:        data,
		serviceName: serviceName,
	}
}
