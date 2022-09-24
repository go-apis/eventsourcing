package providers

import (
	"github.com/contextcloud/eventstore/es/providers/data"
	"github.com/contextcloud/eventstore/es/providers/stream"
)

type Config struct {
	ServiceName string
	Version     string

	Data   data.Config
	Stream stream.Config
}
