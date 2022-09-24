package gpub

import (
	"context"

	"cloud.google.com/go/pubsub"
	"github.com/contextcloud/eventstore/pkg/pub"
)

type Subscription struct {
	sub *pubsub.Subscription
}

func NewSubscription(sub *pubsub.Subscription) *Subscription {
	return &Subscription{
		sub: sub,
	}
}

func (s *Subscription) Receive(ctx context.Context, callback pub.Callback) error {
	return s.sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		if err := callback(ctx, msg.Data); err != nil {
			msg.Nack()
			return
		}
		msg.Ack()
	})
}
