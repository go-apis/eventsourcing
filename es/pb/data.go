package pb

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/contextcloud/eventstore/es"
	"github.com/contextcloud/eventstore/es/filters"
	"github.com/contextcloud/eventstore/es/pb/store"
	"github.com/google/uuid"
)

type data struct {
	storeClient store.StoreClient

	transactionId *string
}

func (d *data) Begin(ctx context.Context) (es.Tx, error) {
	if d.transactionId == nil {
		req := &store.NewTxRequest{}
		resp, err := d.storeClient.NewTx(ctx, req)
		if err != nil {
			return nil, err
		}
		d.transactionId = &resp.TransactionId
	}

	return newTransaction(d.storeClient, *d.transactionId), nil
}
func (d *data) LoadSnapshot(ctx context.Context, serviceName string, aggregateName string, namespace string, id uuid.UUID, out es.SourcedAggregate) error {
	return nil
}
func (d *data) GetEventDatas(ctx context.Context, serviceName string, aggregateName string, namespace string, id uuid.UUID, fromVersion int) ([]json.RawMessage, error) {
	f := int64(fromVersion)
	idStr := id.String()

	req := &store.EventsRequest{
		TransactionId: d.transactionId,
		ServiceName:   &serviceName,
		AggregateType: &aggregateName,
		AggregateId:   &idStr,
		Namespace:     &namespace,
		FromVersion:   &f,
	}
	resp, err := d.storeClient.Events(ctx, req)
	if err != nil {
		return nil, err
	}

	var datas []json.RawMessage
	for _, event := range resp.Events {
		datas = append(datas, event.Data)
	}
	return datas, nil
}
func (d *data) SaveEvents(ctx context.Context, events []es.Event) error {
	if d.transactionId == nil {
		return fmt.Errorf("transaction not started")
	}

	evts := make([]*store.Event, len(events))
	for i, event := range events {
		// TODO what about a codec or something?
		data, err := json.Marshal(event.Data)
		if err != nil {
			return err
		}

		evts[i] = &store.Event{
			ServiceName:   event.ServiceName,
			AggregateType: event.AggregateType,
			AggregateId:   event.AggregateId.String(),
			Namespace:     event.Namespace,
			Version:       int64(event.Version),
			Data:          data,
		}
	}

	req := &store.SaveEventsRequest{
		TransactionId: *d.transactionId,
		Events:        evts,
	}
	if _, err := d.storeClient.SaveEvents(ctx, req); err != nil {
		return err
	}
	return nil
}
func (d *data) SaveEntity(ctx context.Context, entity es.Entity) error {
	return nil
}

func (d *data) Load(ctx context.Context, serviceName string, aggregateName string, namespace string, id uuid.UUID, out interface{}) error {
	return nil
}
func (d *data) Find(ctx context.Context, serviceName string, aggregateName string, namespace string, filter filters.Filter, out interface{}) error {
	return nil
}
func (d *data) Count(ctx context.Context, serviceName string, aggregateName string, namespace string, filter filters.Filter) (int, error) {
	return 0, nil
}

func newData(storeClient store.StoreClient) (es.Data, error) {
	return &data{
		storeClient: storeClient,
	}, nil
}
