package es

import (
	"context"
	"fmt"
	"time"

	"github.com/contextcloud/eventstore/es/filters"
	"github.com/google/uuid"
)

type scheduled struct {
	id            uuid.UUID
	ctx           context.Context
	cmd           Command
	executedAfter time.Time
}

type ScheduledCommandNotifier struct {
	C       <-chan time.Time // The channel on which the ticks are delivered.
	stopper func()
}

func (n *ScheduledCommandNotifier) Stop() {
	n.stopper()
}

func NewScheduledCommandNotifier(c chan time.Time, stopper func()) *ScheduledCommandNotifier {
	return &ScheduledCommandNotifier{
		C:       c,
		stopper: stopper,
	}
}

type CommandScheduler interface {
	Errors() <-chan error
	Close(ctx context.Context) error
}

type commandScheduler struct {
	registry Registry
	data     Data
	handler  CommandHandler

	errCh chan error
}

// Errors returns an error channel that will receive errors from handling of
// scheduled commands.
func (c *commandScheduler) Errors() <-chan error {
	return c.errCh
}

// Close closes the command scheduler.
func (c *commandScheduler) Close(ctx context.Context) error {
	return nil
}

func (c *commandScheduler) handle(ctx context.Context, t time.Time) error {
	lock, err := c.data.Lock(ctx)
	if err != nil {
		return err
	}
	defer lock.Unlock(ctx)

	filter := filters.Filter{
		Where: filters.WhereClause{
			Column: "execute_after",
			Op:     filters.OpLessThan,
			Args:   t,
		},
	}
	persistedCommands, err := c.data.FindPersistedCommands(ctx, filter)
	if err != nil {
		return err
	}

	for _, persistedCommand := range persistedCommands {
		if err := c.handler.HandleCommand(ctx, persistedCommand.Command); err != nil {
			return err
		}
		if err := c.data.DeletePersistedCommand(ctx, persistedCommand); err != nil {
			return err
		}
	}

	return nil
}

func (c *commandScheduler) run(ctx context.Context) {
	notifier, err := c.data.NewScheduledCommandNotifier(ctx)
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
			fmt.Printf("scheduled command at %v\n", t)
			if err := c.handle(ctx, t); err != nil {
				c.errCh <- err
			}
		}
	}
}

func NewCommandScheduler(ctx context.Context, registry Registry, conn Conn, handler CommandHandler) (CommandScheduler, error) {
	data, err := conn.NewData(ctx)
	if err != nil {
		return nil, err
	}

	c := &commandScheduler{
		registry: registry,
		data:     data,
		handler:  handler,
		errCh:    make(chan error, 100),
	}
	go c.run(ctx)
	return c, nil
}
