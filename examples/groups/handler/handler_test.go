package handler

import (
	"context"
	"testing"

	"github.com/contextcloud/graceful/config"
)

func Test_It(t *testing.T) {
	h, err := NewHandler(context.Background(), &config.Config{})
	if err != nil {
		t.Error(err)
		return
	}

	t.Log(h)
}
