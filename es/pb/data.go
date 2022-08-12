package pb

import (
	"context"
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

func (d *data) LoadSnapshot(ctx context.Context, serviceName string, aggregateName string, namespace string, revision string, id uuid.UUID, out es.AggregateSourced) error {
	return nil
}
func (d *data) SaveSnapshot(ctx context.Context, serviceName string, aggregateName string, namespace string, revision string, id uuid.UUID, out es.AggregateSourced) error {
	return nil
}

func (d *data) GetEventDatas(ctx context.Context, serviceName string, aggregateName string, namespace string, id uuid.UUID, fromVersion int) ([]*es.EventData, error) {
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

	var datas []*es.EventData
	for _, event := range resp.Events {
		datas = append(datas, &es.EventData{
			Type:    event.Type,
			Data:    event.Data,
			Version: int(event.Version),
		})
	}
	return datas, nil
}
func (d *data) SaveEventDatas(ctx context.Context, serviceName string, aggregateName string, namespace string, id uuid.UUID, datas []*es.EventData) error {
	if d.transactionId == nil {
		return fmt.Errorf("transaction not started")
	}

	return nil
}
func (d *data) SaveEntity(ctx context.Context, serviceName string, aggregateName string, entity es.Entity) error {
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
