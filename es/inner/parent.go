package inner

import (
	"context"

	"github.com/contextcloud/eventstore/es"
)

type parent struct {
}

func (p *parent) Handle(ctx context.Context, evt *es.Event) ([]es.Command, error) {
	return nil, nil
}

func NewParent() *parent {
	return &parent{}
}
