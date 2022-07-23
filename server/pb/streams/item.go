package streams

import "github.com/contextcloud/eventstore/server/pb/store"

type StreamItem struct {
	Stream      store.Store_EventStreamServer
	ServiceName string
	EventTypes  []string
}

func NewStreamItem(stream store.Store_EventStreamServer, serviceName string, eventTypes []string) *StreamItem {
	return &StreamItem{
		Stream:      stream,
		ServiceName: serviceName,
		EventTypes:  eventTypes,
	}
}
