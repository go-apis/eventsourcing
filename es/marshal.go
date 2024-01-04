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
