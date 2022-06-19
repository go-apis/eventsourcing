package local

import (
	"context"
	"eventstore/es"
)

type client struct {
	data        es.Data
	serviceName string
}

func (c *client) NewTx(ctx context.Context) (es.Tx, error) {
	return c.data.NewTx(ctx)
}

func NewClient(data es.Data, serviceName string) es.Client {
	return &client{
		data:        data,
		serviceName: serviceName,
	}
}
