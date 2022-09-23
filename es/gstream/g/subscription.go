package g

import (
	"context"

	"cloud.google.com/go/pubsub"
	"github.com/contextcloud/eventstore/es/gstream"
)

type subscription struct {
	sub *pubsub.Subscription
}

func NewSubscription(sub *pubsub.Subscription) *subscription {
	return &subscription{
		sub: sub,
	}
}

func (s *subscription) Receive(ctx context.Context, callback gstream.Callback) error {
	return s.sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		if err := callback(ctx, msg.Data); err != nil {
			msg.Nack()
			return
		}
		msg.Ack()
	})
}
