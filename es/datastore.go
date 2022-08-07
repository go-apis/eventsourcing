package es

import (
	"context"
	"errors"

	"github.com/contextgg/pkg/events"
	"github.com/contextgg/pkg/ns"
	"github.com/google/uuid"
)

func applyEvents(ctx context.Context, aggregate AggregateSourced, originalEvents []events.Event) error {
	aggregateType := aggregate.GetTypeName()

	for _, event := range originalEvents {
		if event.AggregateType != aggregateType {
			return ErrMismatchedEventType
		}

		// lets build the event!
		if err := aggregate.ApplyEvent(ctx, event); err != nil {
			return ApplyEventError{
				Event: event,
				Err:   err,
			}
		}
		aggregate.IncrementVersion()
	}
	return nil
}

type DataStore interface {
	Load(ctx context.Context, id uuid.UUID, opts ...DataLoadOption) (Entity, error)
	Save(ctx context.Context, aggregate Entity) ([]events.Event, error)
}

type store struct {
	data Data
	opts *EntityOptions
}

func (s *store) loadSourced(ctx context.Context, aggregate AggregateSourced, forced bool) (Entity, error) {
	namespace := ns.FromContext(ctx)

	// load up the aggregate
	if s.opts.MinVersionDiff >= 0 && !forced {
		if err := s.data.LoadSnapshot(ctx, namespace, s.opts.Revision, aggregate); err != nil {
			return nil, err
		}
	}
	// load up the events from the DB.
	originalEvents, err := s.data.LoadEvents(ctx, namespace, aggregate.GetId(), aggregate.GetTypeName(), aggregate.GetVersion())
	if err != nil {
		return nil, err
	}
	if err := applyEvents(ctx, aggregate, originalEvents); err != nil {
		return nil, err
	}
	return aggregate, nil
}
func (s *store) loadEntity(ctx context.Context, entity Entity) (Entity, error) {
	namespace := ns.FromContext(ctx)

	if err := s.data.LoadEntity(ctx, namespace, entity); err != nil && !errors.Is(err, ErrNoRows) {
		return nil, err
	}
	return entity, nil
}
func (s *store) saveSourced(ctx context.Context, aggregate AggregateSourced) ([]events.Event, error) {
	namespace := ns.FromContext(ctx)

	originalVersion := aggregate.GetVersion()

	// now save it!.
	events := aggregate.Events()
	if len(events) > 0 {
		if err := s.data.SaveEvents(ctx, namespace, events...); err != nil {
			return nil, err
		}
		aggregate.ClearEvents()

		// Apply the events so we can save the aggregate
		if err := applyEvents(ctx, aggregate, events); err != nil {
			return nil, err
		}
	}

	// save the snapshot!
	diff := aggregate.GetVersion() - originalVersion
	if diff < 0 {
		return nil, ErrWrongVersion
	}

	if s.opts.MinVersionDiff >= 0 && diff >= s.opts.MinVersionDiff {
		if err := s.data.SaveSnapshot(ctx, namespace, s.opts.Revision, aggregate); err != nil {
			return nil, err
		}
	}

	if s.opts.Project {
		if err := s.data.SaveEntity(ctx, namespace, aggregate); err != nil {
			return nil, err
		}
	}

	return events, nil
}
func (s *store) saveAggregateHolder(ctx context.Context, aggregate AggregateHolder) ([]events.Event, error) {
	namespace := ns.FromContext(ctx)

	if err := s.data.SaveEntity(ctx, namespace, aggregate); err != nil {
		return nil, err
	}

	events := aggregate.EventsToPublish()
	aggregate.ClearEvents()
	return events, nil
}
func (s *store) saveEntity(ctx context.Context, aggregate Entity) ([]events.Event, error) {
	namespace := ns.FromContext(ctx)
	return nil, s.data.SaveEntity(ctx, namespace, aggregate)
}

func (s *store) Load(ctx context.Context, id string, opts ...DataLoadOption) (Entity, error) {
	options := &DataLoadOptions{}
	for _, o := range opts {
		o(options)
	}

	// make the aggregate
	entity, err := s.opts.Factory()
	if err != nil {
		return nil, err
	}

	switch agg := entity.(type) {
	case AggregateSourced:
		return s.loadSourced(ctx, agg, options.Force)
	default:
		return s.loadEntity(ctx, agg)
	}
}
func (s *store) Save(ctx context.Context, entity Entity) ([]events.Event, error) {
	switch agg := entity.(type) {
	case AggregateSourced:
		return s.saveSourced(ctx, agg)
	case AggregateHolder:
		return s.saveAggregateHolder(ctx, agg)
	default:
		return s.saveEntity(ctx, agg)
	}
}

// NewDataStore for creating stores
func NewDataStore(data Data, opts *EntityOptions) DataStore {
	s := &store{
		data: data,
		opts: opts,
	}
	return s
}
