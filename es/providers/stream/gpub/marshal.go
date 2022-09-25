package gpub

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/contextcloud/eventstore/es"
	"go.opentelemetry.io/otel"
)

func MarshalEvent(ctx context.Context, event *es.Event) ([]byte, error) {
	_, span := otel.Tracer("local").Start(ctx, "MarshalEvent")
	defer span.End()

	// Marshal the event (using JSON for now).
	b, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("could not marshal event: %w", err)
	}

	return b, nil
}

func UnmarshalEvent(ctx context.Context, mappers map[string]es.EventDataFunc, b []byte) (*es.Event, error) {
	_, span := otel.Tracer("local").Start(ctx, "UnmarshalEvent")
	defer span.End()

	out := struct {
		*es.Event

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

	evt := &es.Event{
		ServiceName:   out.ServiceName,
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

	return evt, nil
}
