package es

import (
	"testing"
)

func Test_ProjectorHandle(t *testing.T) {
	projector := &demoProjector{}
	handles := FindProjectorHandles(projector)

	if len(handles) != 1 {
		t.Errorf("expected 1 handle, got %d", len(handles))
	}
}
