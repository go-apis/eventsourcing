package tests

import (
	"testing"

	store "github.com/contextcloud/eventstore/server/pb/store"
)

func TestServer(t *testing.T) {
	api, err := CreateApiServer()
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	stream := NewStreamMock()
	go func() {
		err := api.EventStream(stream)
		if err != nil {
			t.Errorf(err.Error())
		}
		close(stream.sentFromServer)
		close(stream.recvToServer)
	}()

	if err := stream.SendFromClient(&store.EventStreamRequest{}); err != nil {
		t.Errorf(err.Error())
		return
	}

	sumStreamResponse, err := stream.RecvToClient()
	if err != nil {
		t.Error(err.Error())
		return
	}
	if sumStreamResponse.Event == nil {
		t.Errorf("events is nil")
	}
}
