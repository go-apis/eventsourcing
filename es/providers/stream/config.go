package stream

import "github.com/contextcloud/eventstore/pkg/pub"

type Config struct {
	Type   string
	Stream *pub.Config
}
