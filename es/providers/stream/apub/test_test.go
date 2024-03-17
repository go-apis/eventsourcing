package apub

import (
	"context"
	"testing"
	"time"

	"github.com/contextcloud/eventstore/es"
	"github.com/google/uuid"
)

type FakeData struct {
	es.BaseEvent `es:"test"`

	Test string `json:"test"`
}

type FakeHandler struct {
	es.BaseEventHandler `es:"group=external"`

	Key     string
	Results []*es.Event
}

func (f *FakeHandler) HandleEvent(ctx context.Context, evt *es.Event, data *FakeData) error {
	f.Results = append(f.Results, evt)
	return nil
}

type FakeMessageHandler struct {
}

func (f *FakeMessageHandler) HandleGroupMessage(ctx context.Context, group string, msg []byte) error {
	return nil
}

func TestIt(t *testing.T) {
	ctx := context.Background()
	service := "tester"
	snsCfg := &es.AwsSnsConfig{
		Profile:   "Production",
		TopicArn:  "arn:aws:sns:us-east-1:211125614781:prod-events.fifo",
		QueueName: "noops-prod-identity-events.fifo",
	}
	evt1 := &es.Event{
		Service:       service,
		Namespace:     "test",
		Type:          "test",
		AggregateId:   uuid.New(),
		AggregateType: "test",
		Version:       1,
		Timestamp:     time.Now(),
		Data:          &FakeData{Test: "test"},
		Metadata:      make(map[string]interface{}),
	}
	evt2 := &es.Event{
		Service:       service,
		Namespace:     "test",
		Type:          "test",
		AggregateId:   uuid.New(),
		AggregateType: "test",
		Version:       1,
		Timestamp:     time.Now(),
		Data:          &FakeData{Test: "test"},
		Metadata:      make(map[string]interface{}),
	}

	reg, err := es.NewRegistry(service, &FakeData{}, &FakeHandler{})
	if err != nil {
		t.Fatal(err)
		return
	}

	messageHandler := &FakeMessageHandler{}

	streamer, err := NewStreamer(ctx, service, snsCfg, reg, messageHandler)
	if err != nil {
		t.Fatal(err)
		return
	}
	defer func() {
		if err := streamer.Close(ctx); err != nil {
			t.Fatal(err)
		}
	}()

	// publish it.
	if err := streamer.Publish(ctx, evt1); err != nil {
		t.Fatal(err)
		return
	}
	if err := streamer.Publish(ctx, evt2); err != nil {
		t.Fatal(err)
		return
	}

	// wait for it.
	time.Sleep(1 * time.Minute)
}
