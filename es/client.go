package es

import "context"

type Client interface {
	NewUnit(ctx context.Context) (Unit, error)
}

type client struct {
	cfg  Config
	conn Conn
}

func (c *client) NewUnit(ctx context.Context) (Unit, error) {
	data, err := c.conn.NewData(ctx)
	if err != nil {
		return nil, err
	}

	return newUnit(c.cfg, data)
}

func NewClient(cfg Config, conn Conn) (Client, error) {
	ctx := context.Background()
	if err := conn.Initialize(ctx, cfg); err != nil {
		return nil, err
	}

	cli := &client{
		cfg:  cfg,
		conn: conn,
	}
	return cli, nil
}
