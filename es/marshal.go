package es

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

type ServiceEvent struct {
	*Event

	Service string `json:"service"`
}

func MarshalEvent(ctx context.Context, service string, event *Event) ([]byte, error) {
	d := &ServiceEvent{
		Event:   event,
		Service: service,
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
		return nil, fmt.Errorf("could not decode event: %w", err)
	}

	evtConfig, ok := cfg.GetEventConfigs()[strings.ToLower(out.Type)]
	if !ok {
		return nil, nil
	}

	service := cfg.GetProviderConfig().Service
	if strings.EqualFold(service, out.Service) {
		return nil, nil
	}

	if evtConfig.Service != nil && *evtConfig.Service != out.Service {
		return nil, nil
	}

	data, err := evtConfig.Factory()
	if err != nil {
		return nil, fmt.Errorf("could not create event: %w", err)
	}

	if err := json.Unmarshal(out.Data, data); err != nil {
		return nil, fmt.Errorf("could not decode event: %w", err)
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
		Event:   evt,
		Service: out.Service,
	}

	return with, nil
}
