package es

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

func applyEvents(ctx context.Context, aggregate AggregateSourced, datas []*EventData) error {
	for _, d := range datas {
		if err := json.Unmarshal(d.Data, aggregate); err != nil {
			return err
		}
		aggregate.IncrementVersion()
	}
	return nil
}

type DataStore interface {
	Load(ctx context.Context, id uuid.UUID, opts ...DataLoadOption) (Entity, error)
	Save(ctx context.Context, aggregate Entity) ([]Event, error)
}

type store struct {
	serviceName string
	data        Data
	opts        *EntityOptions
}

func (s *store) loadSourced(ctx context.Context, aggregate AggregateSourced, forced bool) (Entity, error) {
	namespace := NamespaceFromContext(ctx)
	id := aggregate.GetId()

	// load up the aggregate
	if s.opts.MinVersionDiff >= 0 && !forced {
		if err := s.data.LoadSnapshot(ctx, s.serviceName, s.opts.Name, namespace, s.opts.Revision, id, aggregate); err != nil {
			return nil, err
		}
	}
	// load up the events from the DB.
	originalEvents, err := s.data.GetEventDatas(ctx, s.serviceName, s.opts.Name, namespace, id, aggregate.GetVersion())
	if err != nil {
		return nil, err
	}
	if err := applyEvents(ctx, aggregate, originalEvents); err != nil {
		return nil, err
	}
	return aggregate, nil
}
func (s *store) loadEntity(ctx context.Context, entity Entity) (Entity, error) {
	namespace := NamespaceFromContext(ctx)

	id := entity.GetId()
	if err := s.data.Load(ctx, s.serviceName, s.opts.Name, namespace, id, entity); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return entity, nil
}
func (s *store) saveSourced(ctx context.Context, aggregate AggregateSourced) ([]Event, error) {
	namespace := NamespaceFromContext(ctx)
	version := aggregate.GetVersion()
	id := aggregate.GetId()
	raw := aggregate.GetEvents()
	if len(raw) > 0 {
		datas := make([]*EventData, len(raw))
		for i, data := range raw {
			name := fmt.Sprintf("%T", data)
			v := version + i + 1
			d, err := json.Marshal(data)
			if err != nil {
				return nil, err
			}

			datas[i] = &EventData{
				Type:    name,
				Version: v,
				Data:    d,
			}
		}

		if err := s.data.SaveEventDatas(ctx, s.serviceName, s.opts.Name, namespace, id, datas); err != nil {
			return nil, err
		}

		// Apply the events so we can save the aggregate
		if err := applyEvents(ctx, aggregate, datas); err != nil {
			return nil, err
		}
	}

	// save the snapshot!
	diff := aggregate.GetVersion() - version
	if diff < 0 {
		return nil, fmt.Errorf("version diff is less than 0")
	}

	if s.opts.MinVersionDiff >= 0 && diff >= s.opts.MinVersionDiff {
		if err := s.data.SaveSnapshot(ctx, s.serviceName, s.opts.Name, namespace, s.opts.Revision, id, aggregate); err != nil {
			return nil, err
		}
	}

	if s.opts.Project {
		if err := s.data.SaveEntity(ctx, s.serviceName, s.opts.Name, aggregate); err != nil {
			return nil, err
		}
	}

	return nil, nil
}
func (s *store) saveAggregateHolder(ctx context.Context, aggregate AggregateHolder) ([]Event, error) {
	if err := s.data.SaveEntity(ctx, s.serviceName, s.opts.Name, aggregate); err != nil {
		return nil, err
	}

	events := aggregate.EventsToPublish()
	return events, nil
}
func (s *store) saveEntity(ctx context.Context, aggregate Entity) ([]Event, error) {
	return nil, s.data.SaveEntity(ctx, s.serviceName, s.opts.Name, aggregate)
}

func (s *store) Load(ctx context.Context, id uuid.UUID, opts ...DataLoadOption) (Entity, error) {
	options := &DataLoadOptions{}
	for _, o := range opts {
		o(options)
	}
	namespace := NamespaceFromContext(ctx)

	// make the aggregate
	entity, err := s.opts.Factory()
	if err != nil {
		return nil, err
	}

	switch agg := entity.(type) {
	case SetId:
		agg.SetId(id, namespace)
	}

	switch agg := entity.(type) {
	case AggregateSourced:
		return s.loadSourced(ctx, agg, options.Force)
	default:
		return s.loadEntity(ctx, agg)
	}
}
func (s *store) Save(ctx context.Context, entity Entity) ([]Event, error) {
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
func NewDataStore(serviceName string, data Data, opts *EntityOptions) DataStore {
	s := &store{
		serviceName: serviceName,
		data:        data,
		opts:        opts,
	}
	return s
}
