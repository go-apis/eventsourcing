package es

import (
	"context"
	"encoding/json"
	"fmt"
)

func MarshalEvent(ctx context.Context, event *Event) ([]byte, error) {
	// Marshal the event (using JSON for now).
	b, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("could not marshal event: %w", err)
	}

	return b, nil
}

func UnmarshalEvent(ctx context.Context, b []byte) (*Event, error) {
	out := struct {
		*Event

		Data json.RawMessage `json:"data"`
	}{}

	if err := json.Unmarshal(b, &out); err != nil {
		return nil, fmt.Errorf("could not decode event: %w", err)
	}

	evtConfig, err := GlobalRegistry.GetEventConfig(out.Service, out.Type)
	if err != nil {
		return nil, err
	}

	data, err := evtConfig.Factory()
	if err != nil {
		return nil, fmt.Errorf("could not create event: %w", err)
	}

	if err := json.Unmarshal(out.Data, data); err != nil {
		return nil, fmt.Errorf("could not decode event: %w", err)
	}

	metadata := out.Metadata
	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	return &Event{
		Service:       out.Service,
		Namespace:     out.Namespace,
		AggregateId:   out.AggregateId,
		AggregateType: out.AggregateType,
		Version:       out.Version,
		Type:          out.Type,
		Timestamp:     out.Timestamp,
		Metadata:      metadata,
		Data:          data,
	}, nil
}
