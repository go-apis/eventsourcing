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
	Key     string
	Results []*es.Event
}

func (f *FakeHandler) Handle(ctx context.Context, evt *es.Event) error {
	f.Results = append(f.Results, evt)
	return nil
}

func TestIt(t *testing.T) {
	ctx := context.Background()
	service := "tester"
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

	reg, err := es.NewRegistry(service, &FakeData{})
	if err != nil {
		t.Fatal(err)
		return
	}

	fakeHandler1 := &FakeHandler{}
	fakeHandler2 := &FakeHandler{
		Key: "test2",
	}

	streamer, err := NewStreamer(ctx, service, &es.AwsSnsConfig{
		Profile:  "Development",
		Region:   "us-east-1",
		TopicArn: "arn:aws:sns:us-east-1:888821167166:deployment.fifo",
	}, reg.ParseEvent)
	if err != nil {
		t.Fatal(err)
		return
	}
	defer func() {
		if err := streamer.Close(ctx); err != nil {
			t.Fatal(err)
		}
	}()

	if err := streamer.AddHandler(ctx, fakeHandler1.Key, fakeHandler1); err != nil {
		t.Fatal(err)
		return
	}
	if err := streamer.AddHandler(ctx, fakeHandler2.Key, fakeHandler2); err != nil {
		t.Fatal(err)
		return
	}

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

	// check it.
	if len(fakeHandler1.Results) != 2 {
		t.Fatalf("expected 2 events, got %d", len(fakeHandler1.Results))
		return
	}

	if len(fakeHandler2.Results) != 2 {
		t.Fatalf("expected 2 events, got %d", len(fakeHandler2.Results))
		return
	}
}
