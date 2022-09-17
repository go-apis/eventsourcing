package api

import (
	"context"
	"net/http"

	"github.com/contextcloud/graceful/config"
)

func NewHandler(ctx context.Context, cfg *config.Config) (http.Handler, error) {
	return nil, nil
}
