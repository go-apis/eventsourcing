package apub

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"testing"
	"time"

	"github.com/go-apis/eventsourcing/es"
	"github.com/google/uuid"
)

type FakeData struct {
	es.BaseEvent `es:"test"`

	Test string `json:"test"`
}

type FakeHandler struct {
	es.BaseEventHandler `es:"group=external"`
}

func (f *FakeHandler) HandleEvent(ctx context.Context, evt *es.Event, data *FakeData) error {
	return nil
}

type FakeMessageHandler struct {
}

func (f *FakeMessageHandler) HandleGroupMessage(ctx context.Context, group string, msg []byte) error {
	fmt.Printf("group: %s, message: %s\n", group, string(msg))
	return nil
}

func TestIt(t *testing.T) {
	ctx := context.Background()
	service := "tester"
	snsCfg := &es.AwsSnsConfig{
		Profile:   "Test",
		Region:    "us-east-1",
		TopicArn:  "arn:aws:sns:us-east-1:211125614781:prod-events-test.fifo",
		QueueName: "noops-prod-test-events.fifo",
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

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		go func() {
			<-c
			os.Exit(1)
		}()
	}()

	for {
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
		// publish it.
		if err := streamer.Publish(ctx, evt1); err != nil {
			t.Fatal(err)
			return
		}
		time.Sleep(5 * time.Second)
	}
}
