package tests

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/contextcloud/eventstore/es"
	"github.com/contextcloud/eventstore/es/providers/stream/gpub"
	"github.com/contextcloud/eventstore/pkg/gcppubsub"
	"github.com/google/uuid"
)

type MyEvent struct {
	Name string
}

func Test_It(t *testing.T) {
	testEvents := []*es.Event{
		{
			Namespace:     "default",
			AggregateId:   uuid.New(),
			AggregateType: "MyAggregate",
			Version:       1,
			Type:          "MyEvent",
			Timestamp:     time.Now(),
			Data:          &MyEvent{Name: "test"},
		},
	}

	cfg := &gcppubsub.Config{
		ProjectId: "nordic-gaming",
		TopicId:   "test_topic",
	}
	if err := gcppubsub.Reset(cfg); err != nil {
		t.Error(err)
		return
	}

	cli, err := gcppubsub.Open(cfg)
	if err != nil {
		t.Error(err)
		return
	}

	streamer, err := gpub.NewStreamer(cli)
	if err != nil {
		t.Error(err)
		return
	}

	initOpts := es.InitializeOptions{
		ServiceName: "demo",
		EventConfigs: []*es.EventConfig{
			{
				Name: "MyEvent",
				Type: reflect.TypeOf(&MyEvent{}),
				Factory: func() (interface{}, error) {
					return &MyEvent{}, nil
				},
			},
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var out []*es.Event

	callback := func(ctx context.Context, evt *es.Event) error {
		out = append(out, evt)
		t.Logf("got event: %+v", evt)

		if len(out) == len(testEvents) {
			cancel()
		}
		return nil
	}

	if err := streamer.Start(ctx, initOpts, callback); err != nil {
		t.Error(err)
		return
	}

	if err := streamer.Publish(ctx, testEvents...); err != nil {
		t.Error(err)
		return
	}

	<-ctx.Done()

	if len(out) != len(testEvents) {
		t.Errorf("expected %d events, got %d", len(testEvents), len(out))
		return
	}
}
