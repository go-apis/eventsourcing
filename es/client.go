package es

import "context"

type Client interface {
	NewUnit(ctx context.Context) (Unit, error)
}

type client struct {
	cfg  Config
	data Data
}

func (c *client) NewUnit(ctx context.Context) (Unit, error) {
	tx, err := c.data.NewTx(ctx)
	if err != nil {
		return nil, err
	}
	return NewUnit(c.cfg, tx)
}

func NewClient(cfg Config, data Data) Client {
	cli := &client{
		cfg:  cfg,
		data: data,
	}
	return cli
}
