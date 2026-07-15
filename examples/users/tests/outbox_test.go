package tests

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/go-apis/eventsourcing/es"
	"github.com/go-apis/eventsourcing/examples/users/data/commands"
	"github.com/go-apis/eventsourcing/examples/users/helpers"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// wireEvent is the subset of the published payload the tests assert on.
type wireEvent struct {
	Service       string    `json:"service"`
	AggregateId   uuid.UUID `json:"aggregate_id"`
	AggregateType string    `json:"aggregate_type"`
	Version       int       `json:"version"`
	Type          string    `json:"type"`
}

// receive waits for the next message on the bus and acks it.
func receive(t *testing.T, messages <-chan *message.Message, timeout time.Duration) *wireEvent {
	t.Helper()

	select {
	case msg, ok := <-messages:
		require.True(t, ok, "bus subscription closed")
		defer msg.Ack()

		var evt wireEvent
		require.NoError(t, json.Unmarshal(msg.Payload, &evt))
		return &evt
	case <-time.After(timeout):
		return nil
	}
}

// outboxDepth counts pending outbox rows through a claim in a rolled-back
// transaction.
func outboxDepth(t *testing.T, cli es.Client) int {
	t.Helper()

	ctx := context.Background()
	unit, err := cli.Unit(ctx)
	require.NoError(t, err)

	data := unit.Data()
	tx, err := data.Begin(ctx)
	require.NoError(t, err)
	defer func() { _ = tx.Rollback(ctx) }()

	rows, err := data.ClaimOutbox(ctx, 1000)
	require.NoError(t, err)
	return len(rows)
}

func requireOutboxDrained(t *testing.T, cli es.Client) {
	t.Helper()

	require.Eventually(t, func() bool {
		return outboxDepth(t, cli) == 0
	}, 5*time.Second, 50*time.Millisecond, "outbox should drain to empty")
}

// TestOutboxDeliversToBus is the happy path: a dispatched command whose
// event is publish-tagged reaches the bus via the outbox relay, and the
// outbox drains back to empty.
func TestOutboxDeliversToBus(t *testing.T) {
	tester, err := NewTester()
	require.NoError(t, err)

	messages, err := tester.PubSub().Subscribe(context.Background(), "test")
	require.NoError(t, err)

	cli := tester.Client()
	ctx := context.Background()
	unit, err := cli.Unit(ctx)
	require.NoError(t, err)
	ctx = es.SetUnit(ctx, unit)
	ctx = helpers.SetSkipSaga(ctx)

	userId := uuid.New()
	groupId := uuid.New()
	require.NoError(t, unit.Dispatch(ctx,
		&commands.CreateUser{
			BaseCommand: es.BaseCommand{AggregateId: userId},
			Username:    "outbox.tester",
			Password:    "12345678",
		},
		// GroupAdded is the example's only publish-tagged event.
		&commands.AddGroup{
			BaseCommand: es.BaseCommand{AggregateId: userId},
			GroupId:     groupId,
		},
	))

	evt := receive(t, messages, 3*time.Second)
	require.NotNil(t, evt, "expected the published event on the bus")
	require.Equal(t, "GroupAdded", evt.Type)
	require.Equal(t, userId, evt.AggregateId)

	requireOutboxDrained(t, cli)
}

// TestOutboxSkipPublish: SkipPublish suppresses outbox rows entirely, so
// nothing reaches the bus.
func TestOutboxSkipPublish(t *testing.T) {
	tester, err := NewTester()
	require.NoError(t, err)

	messages, err := tester.PubSub().Subscribe(context.Background(), "test")
	require.NoError(t, err)

	cli := tester.Client()
	ctx := es.SetSkipPublish(context.Background())
	unit, err := cli.Unit(ctx)
	require.NoError(t, err)
	ctx = es.SetUnit(ctx, unit)
	ctx = helpers.SetSkipSaga(ctx)

	userId := uuid.New()
	require.NoError(t, unit.Dispatch(ctx,
		&commands.CreateUser{
			BaseCommand: es.BaseCommand{AggregateId: userId},
			Username:    "outbox.skipped",
			Password:    "12345678",
		},
		&commands.AddGroup{
			BaseCommand: es.BaseCommand{AggregateId: userId},
			GroupId:     uuid.New(),
		},
	))

	require.Equal(t, 0, outboxDepth(t, cli), "skip-publish must not park outbox rows")
	evt := receive(t, messages, 500*time.Millisecond)
	require.Nil(t, evt, "skip-publish must not publish to the bus")
}

// TestOutboxResumeDrainsPending is the crash-recovery path: rows already in
// the outbox that nobody nudges the relay about (as left behind by a killed
// process) are picked up by the poll loop and published.
func TestOutboxResumeDrainsPending(t *testing.T) {
	restore := es.OutboxPollInterval
	es.OutboxPollInterval = 100 * time.Millisecond
	defer func() { es.OutboxPollInterval = restore }()

	tester, err := NewTester()
	require.NoError(t, err)

	messages, err := tester.PubSub().Subscribe(context.Background(), "test")
	require.NoError(t, err)

	cli := tester.Client()
	ctx := context.Background()
	unit, err := cli.Unit(ctx)
	require.NoError(t, err)

	// Park a row directly, bypassing the unit's nudge — exactly the state a
	// process killed after commit but before publish leaves behind.
	evt := &es.Event{
		Service:       "users",
		Namespace:     "default",
		AggregateId:   uuid.New(),
		AggregateType: "StandardUser",
		Version:       1,
		Type:          "GroupAdded",
		Timestamp:     time.Now(),
		Data:          map[string]any{},
	}
	payload, err := es.MarshalEvent(ctx, evt)
	require.NoError(t, err)

	data := unit.Data()
	tx, err := data.Begin(ctx)
	require.NoError(t, err)
	require.NoError(t, data.SaveOutbox(ctx, []*es.OutboxEvent{{
		Service:     "users",
		OrderingKey: es.EventOrderingKey(evt),
		Payload:     payload,
	}}))
	require.NoError(t, tx.Commit(ctx))

	got := receive(t, messages, 3*time.Second)
	require.NotNil(t, got, "poll loop should publish rows left behind by a crash")
	require.Equal(t, "GroupAdded", got.Type)
	require.Equal(t, evt.AggregateId, got.AggregateId)

	requireOutboxDrained(t, cli)
}

// TestDetachedUnit: a detached unit ignores the ctx unit and commits on its
// own, so handlers can land chunk increments independently of the
// delivery-wide transaction.
func TestDetachedUnit(t *testing.T) {
	tester, err := NewTester()
	require.NoError(t, err)

	cli := tester.Client()
	ctx := es.SetClient(context.Background(), cli)

	outer, err := cli.Unit(ctx)
	require.NoError(t, err)
	ctx = es.SetUnit(ctx, outer)

	detached, err := es.NewDetachedUnit(ctx)
	require.NoError(t, err)
	require.NotSame(t, outer, detached)

	dctx := helpers.SetSkipSaga(es.SetUnit(ctx, detached))
	userId := uuid.New()
	require.NoError(t, detached.Dispatch(dctx,
		&commands.CreateUser{
			BaseCommand: es.BaseCommand{AggregateId: userId},
			Username:    "detached.tester",
			Password:    "12345678",
		},
	))

	// Committed by the detached unit: visible to an unrelated fresh unit.
	freshCtx := context.Background()
	fresh, err := cli.Unit(freshCtx)
	require.NoError(t, err)
	freshCtx = es.SetUnit(freshCtx, fresh)
	var out struct {
		Id       uuid.UUID `json:"id"`
		Username string    `json:"username"`
	}
	require.NoError(t, fresh.Get(freshCtx, "StandardUser", "default", userId, &out))
	require.Equal(t, "detached.tester", out.Username)
}
