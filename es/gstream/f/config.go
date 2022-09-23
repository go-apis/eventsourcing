package f

import (
	"fmt"

	"github.com/contextcloud/eventstore/es/gstream"
	"github.com/contextcloud/eventstore/es/gstream/g"
)

type Config struct {
	Type      string
	Url       string
	ProjectId string
	TopicId   string
}

func NewClient(cfg *Config) (gstream.Client, error) {
	switch cfg.Type {
	case "gcp":
		return g.Open(
			g.WithProjectId(cfg.ProjectId),
			g.WithTopicId(cfg.TopicId),
		)
	default:
		return nil, fmt.Errorf("unknown type: %s", cfg.Type)
	}
}
