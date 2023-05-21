package es

import (
	"context"
	"encoding/json"
	"fmt"
)

type EventPublished struct {
	*Event

	ServiceName string `json:"service_name"`
}

func MarshalEvent(ctx context.Context, serviceName string, event *Event) ([]byte, error) {
	d := &EventPublished{
		Event:       event,
		ServiceName: serviceName,
	}

	// Marshal the event (using JSON for now).
	b, err := json.Marshal(d)
	if err != nil {
		return nil, fmt.Errorf("could not marshal event: %w", err)
	}

	return b, nil
}

func UnmarshalEvent(ctx context.Context, mappers map[string]EventDataFunc, b []byte) (*EventPublished, error) {
	out := struct {
		*EventPublished

		Data json.RawMessage `json:"data"`
	}{}

	if err := json.Unmarshal(b, &out); err != nil {
		return nil, fmt.Errorf("Could not decode event: %w", err)
	}

	builder, ok := mappers[out.Type]
	if !ok {
		return nil, nil
	}

	data, err := builder()
	if err != nil {
		return nil, fmt.Errorf("Could not create event: %w", err)
	}

	if err := json.Unmarshal(out.Data, data); err != nil {
		return nil, fmt.Errorf("Could not decode event: %w", err)
	}

	evt := &Event{
		Namespace:     out.Namespace,
		AggregateId:   out.AggregateId,
		AggregateType: out.AggregateType,
		Version:       out.Version,
		Type:          out.Type,
		Timestamp:     out.Timestamp,
		Metadata:      out.Metadata,
		Data:          data,
	}

	if evt.Metadata == nil {
		evt.Metadata = make(map[string]interface{})
	}

	with := &EventPublished{
		Event:       evt,
		ServiceName: out.ServiceName,
	}

	return with, nil
}
