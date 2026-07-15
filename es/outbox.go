package es

import (
	"context"
	"log/slog"
	"strings"
	"sync"
	"time"
)

// Outbox tuning knobs. Package-level so deployments (and tests) can adjust
// them before NewClient; the defaults suit a service with a shared Postgres.
var (
	// OutboxBatchSize is the maximum rows claimed per relay transaction.
	OutboxBatchSize = 500
	// OutboxPollInterval is how often the relay checks for rows nobody
	// nudged it about — the crash-recovery path picks pending rows up
	// within this interval after a restart.
	OutboxPollInterval = 5 * time.Second
	// OutboxPublishConcurrency bounds in-flight publishes across ordering
	// keys; rows sharing an ordering key always publish sequentially.
	OutboxPublishConcurrency = 16
)

// OutboxEvent is a publishable event parked durably in the same transaction
// as the events it belongs to, then relayed to the stream by the outbox
// relay. Payload is the marshaled wire event, frozen at commit time.
type OutboxEvent struct {
	Id          int64
	Service     string
	OrderingKey string
	Payload     []byte
	CreatedAt   time.Time
}

// RawEventPublisher is implemented by streamers that can publish a
// pre-marshaled event without reparsing it. The relay prefers it and falls
// back to registry.ParseEvent + Publish for streamers that don't.
type RawEventPublisher interface {
	PublishRaw(ctx context.Context, orderingKey string, payload []byte) error
}

// PublishNotifier wakes the outbox relay after a unit commits rows. Units
// hold this instead of a direct publisher: delivery to the stream is the
// relay's job.
type PublishNotifier interface {
	Nudge()
}

// outboxRelay drains the outbox table to the streamer. One relay runs per
// client; a per-service advisory lock in ClaimOutbox keeps a single relay
// active across service instances so ordering keys drain in insert order.
// Claim, publish and delete share one transaction — a crash at any point
// rolls the claim back and the rows are re-published later, so delivery is
// at-least-once (which bus consumers must already tolerate).
type outboxRelay struct {
	service   string
	registry  Registry
	publisher EventPublisher
	data      Data

	nudge     chan struct{}
	closed    chan struct{}
	closeOnce sync.Once
}

func newOutboxRelay(ctx context.Context, service string, conn Conn, registry Registry, publisher EventPublisher) (*outboxRelay, error) {
	data, err := conn.NewData(ctx)
	if err != nil {
		return nil, err
	}

	return &outboxRelay{
		service:   service,
		registry:  registry,
		publisher: publisher,
		data:      data,
		nudge:     make(chan struct{}, 1),
		closed:    make(chan struct{}),
	}, nil
}

// Nudge wakes the relay without blocking; a wake-up is already pending if
// the channel is full, which is enough — the relay drains until empty.
func (r *outboxRelay) Nudge() {
	select {
	case r.nudge <- struct{}{}:
	default:
	}
}

func (r *outboxRelay) Close(ctx context.Context) error {
	r.closeOnce.Do(func() {
		close(r.closed)
	})
	return nil
}

func (r *outboxRelay) run(ctx context.Context) {
	// Drain whatever a previous process left behind before steady state —
	// this is the crash-recovery path.
	r.drain(ctx)

	ticker := time.NewTicker(OutboxPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-r.closed:
			return
		case <-r.nudge:
		case <-ticker.C:
		}
		r.drain(ctx)
	}
}

// drain publishes batches until the outbox is empty or a batch fails; a
// failed batch is retried on the next nudge or tick rather than hot-looped.
func (r *outboxRelay) drain(ctx context.Context) {
	for {
		n, err := r.drainBatch(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "outbox drain failed",
				"service", r.service,
				"error", err,
			)
			return
		}
		if n == 0 {
			return
		}
	}
}

func (r *outboxRelay) drainBatch(ctx context.Context) (published int, err error) {
	tx, err := r.data.Begin(ctx)
	if err != nil {
		return 0, err
	}
	committed := false
	defer func() {
		if !committed {
			if rerr := tx.Rollback(ctx); rerr != nil {
				slog.ErrorContext(ctx, "outbox claim rollback failed",
					"service", r.service,
					"error", rerr,
				)
			}
		}
	}()

	rows, err := r.data.ClaimOutbox(ctx, OutboxBatchSize)
	if err != nil {
		return 0, err
	}
	// No rows, or another instance holds the relay lock.
	if len(rows) == 0 {
		return 0, nil
	}

	done, publishErr := r.publish(ctx, rows)
	if len(done) > 0 {
		if derr := r.data.DeleteOutbox(ctx, done); derr != nil {
			return 0, derr
		}
	}
	if cerr := tx.Commit(ctx); cerr != nil {
		return 0, cerr
	}
	committed = true

	return len(done), publishErr
}

// sequenceKey is the per-aggregate grouping the relay serializes on. The
// broker ordering key ends in the event version, so grouping on it directly
// would put every row in its own group and let one aggregate's events
// publish out of order; stripping the version leaves the aggregate identity.
func sequenceKey(orderingKey string) string {
	if i := strings.LastIndex(orderingKey, ":"); i >= 0 {
		return orderingKey[:i]
	}
	return orderingKey
}

// publish delivers the claimed rows: aggregates run concurrently (bounded),
// rows of one aggregate sequentially in claim (insert) order, and a failure
// stops that aggregate so later rows never overtake an unpublished earlier
// one. It returns the ids that made it out and the first error encountered.
func (r *outboxRelay) publish(ctx context.Context, rows []*OutboxEvent) ([]int64, error) {
	// Group by aggregate preserving claim (id) order.
	keys := make([]string, 0, len(rows))
	groups := make(map[string][]*OutboxEvent, len(rows))
	for _, row := range rows {
		key := sequenceKey(row.OrderingKey)
		if _, ok := groups[key]; !ok {
			keys = append(keys, key)
		}
		groups[key] = append(groups[key], row)
	}

	var (
		mu       sync.Mutex
		done     = make([]int64, 0, len(rows))
		firstErr error
	)

	sem := make(chan struct{}, OutboxPublishConcurrency)
	var wg sync.WaitGroup
	for _, key := range keys {
		group := groups[key]

		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()

			for _, row := range group {
				if err := r.publishRow(ctx, row); err != nil {
					mu.Lock()
					if firstErr == nil {
						firstErr = err
					}
					mu.Unlock()
					return
				}
				mu.Lock()
				done = append(done, row.Id)
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	return done, firstErr
}

func (r *outboxRelay) publishRow(ctx context.Context, row *OutboxEvent) error {
	if raw, ok := r.publisher.(RawEventPublisher); ok {
		return raw.PublishRaw(ctx, row.OrderingKey, row.Payload)
	}

	evt, err := r.registry.ParseEvent(ctx, row.Payload)
	if err != nil {
		return err
	}
	return r.publisher.Publish(ctx, evt)
}
