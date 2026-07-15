package es

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

// fakeBatchPublisher fails the message indexes in fail and succeeds the rest.
type fakeBatchPublisher struct {
	fail map[int]bool
}

func (f *fakeBatchPublisher) Publish(ctx context.Context, evt *Event) error {
	return fmt.Errorf("relay must prefer the batch path")
}

func (f *fakeBatchPublisher) PublishRawBatch(ctx context.Context, msgs []RawEvent) []error {
	errs := make([]error, len(msgs))
	for i := range msgs {
		if f.fail[i] {
			errs[i] = fmt.Errorf("publish failed")
		}
	}
	return errs
}

func TestSequenceKey(t *testing.T) {
	require.Equal(t, "ns:aggid:AggType", sequenceKey("ns:aggid:AggType:4"))
	require.Equal(t, "no-colons", sequenceKey("no-colons"))
}

// TestOutboxPublishBatchPrefix: after a failed row, later rows of the same
// aggregate stay pending (no gaps behind a failure), while other aggregates
// are unaffected.
func TestOutboxPublishBatchPrefix(t *testing.T) {
	rows := []*OutboxEvent{
		{Id: 1, OrderingKey: "ns:a:User:1"},
		{Id: 2, OrderingKey: "ns:b:User:1"},
		{Id: 3, OrderingKey: "ns:a:User:2"}, // fails
		{Id: 4, OrderingKey: "ns:a:User:3"}, // broker accepts, must stay pending
		{Id: 5, OrderingKey: "ns:b:User:2"},
	}

	r := &outboxRelay{
		publisher: &fakeBatchPublisher{fail: map[int]bool{2: true}},
	}

	done, err := r.publish(context.Background(), rows)
	require.Error(t, err)
	require.Equal(t, []int64{1, 2, 5}, done)
}
