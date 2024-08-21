package es

import (
	"context"
	"time"
)

type CommandScheduler interface {
	Errors() <-chan error
	Close(ctx context.Context) error
}

type commandScheduler struct {
	cctx   context.Context
	cancel context.CancelFunc

	client *client

	errCh chan error
}

// Errors returns an error channel that will receive errors from handling of
// scheduled commands.
func (c *commandScheduler) Errors() <-chan error {
	return c.errCh
}

// Close closes the command scheduler.
func (c *commandScheduler) Close(ctx context.Context) error {
	c.cancel()
	return nil
}

func (c *commandScheduler) handle(ctx context.Context, t time.Time) error {
	unit, err := c.client.Unit(ctx)
	if err != nil {
		return err
	}

	lock, err := unit.Data().Lock(ctx)
	if err != nil {
		return err
	}
	defer lock.Unlock(ctx)

	filter := Filter{
		Where: WhereClause{
			Column: "execute_after",
			Op:     OpLessThan,
			Args:   t,
		},
	}
	persistedCommands, err := unit.Data().FindPersistedCommands(ctx, filter)
	if err != nil {
		return err
	}

	for _, persistedCommand := range persistedCommands {
		inner := SetActor(ctx, persistedCommand.By)

		if err := unit.Dispatch(inner, persistedCommand.Command); err != nil {
			return err
		}
		if err := unit.Data().DeletePersistedCommand(inner, persistedCommand); err != nil {
			return err
		}
	}

	return nil
}

func (c *commandScheduler) run(ctx context.Context) {
	notifier, err := NewScheduledCommandNotifier(ctx)
	if err != nil {
		c.errCh <- err
		return
	}
	defer notifier.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case t := <-notifier.C:
			if err := c.handle(ctx, t); err != nil {
				c.errCh <- err
			}
		}
	}
}

func NewCommandScheduler(ctx context.Context, client *client) (CommandScheduler, error) {
	cctx, cancel := context.WithCancel(ctx)

	c := &commandScheduler{
		cctx:   cctx,
		cancel: cancel,
		client: client,
		errCh:  make(chan error, 100),
	}
	go c.run(cctx)
	return c, nil
}
