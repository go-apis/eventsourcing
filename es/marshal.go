package es

import (
	"context"
	"encoding/json"
	"fmt"
)

type ServiceEvent struct {
	*Event

	ServiceName string `json:"service_name"`
}

func MarshalEvent(ctx context.Context, serviceName string, event *Event) ([]byte, error) {
	d := &ServiceEvent{
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

func UnmarshalEvent(ctx context.Context, cfg Config, b []byte) (*ServiceEvent, error) {
	out := struct {
		*ServiceEvent

		Data json.RawMessage `json:"data"`
	}{}

	if err := json.Unmarshal(b, &out); err != nil {
		return nil, fmt.Errorf("Could not decode event: %w", err)
	}

	evtConfig, ok := cfg.GetEventConfigs()[out.Type]
	if !ok {
		return nil, nil
	}

	// we only care about service names that match the config
	if evtConfig.ServiceName == nil || *evtConfig.ServiceName != out.ServiceName {
		return nil, nil
	}

	data, err := evtConfig.Factory()
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

	with := &ServiceEvent{
		Event:       evt,
		ServiceName: out.ServiceName,
	}

	return with, nil
}
