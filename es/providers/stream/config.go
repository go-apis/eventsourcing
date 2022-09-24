package stream

import "github.com/contextcloud/eventstore/es/providers/stream/gpub"

type Config struct {
	Type   string
	Stream *gpub.Config
}
