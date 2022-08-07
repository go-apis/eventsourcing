package es

import (
	"context"

	"github.com/google/uuid"
)

type Aggregate interface {
	Apply(ctx context.Context, data interface{}) error
}

type SetId interface {
	SetId(id uuid.UUID, namespace string)
}
