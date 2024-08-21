package es

import (
	"context"
	"time"
)

type ScheduledCommandNotifier struct {
	C       <-chan time.Time // The channel on which the ticks are delivered.
	stopper func()
}

func (n *ScheduledCommandNotifier) Stop() {
	n.stopper()
}

func NewScheduledCommandNotifier(ctx context.Context) (*ScheduledCommandNotifier, error) {
	inner, cancel := context.WithCancel(ctx)

	ch := make(chan time.Time)
	notifier := &ScheduledCommandNotifier{
		C: ch,
		stopper: func() {
			cancel()
			close(ch)
		},
	}

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-inner.Done():
				return
			case t := <-ticker.C:
				ch <- t
			}
		}
	}()

	return notifier, nil
}
