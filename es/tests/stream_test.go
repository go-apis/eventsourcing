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

var _ es.Config = &MyConfig{}

type MyConfig struct {
}

func (MyConfig) GetProviderConfig() *es.ProviderConfig {
	return &es.ProviderConfig{
		ServiceName: "demo",
		Version:     "1.0.0",
	}
}
func (MyConfig) GetEntityConfigs() map[string]*es.EntityConfig {
	return nil
}
func (MyConfig) GetCommandConfigs() map[string]*es.CommandConfig {
	return nil
}
func (MyConfig) GetEventConfigs() map[string]*es.EventConfig {
	return map[string]*es.EventConfig{
		"MyEvent": {
			Name: "MyEvent",
			Type: reflect.TypeOf(&MyEvent{}),
			Factory: func() (interface{}, error) {
				return &MyEvent{}, nil
			},
		},
	}
}
func (MyConfig) GetReplayHandler(entityName string) es.CommandHandler {
	return nil
}
func (MyConfig) GetCommandHandlers() map[reflect.Type]es.CommandHandler {
	return nil
}
func (MyConfig) GetEventHandlers() map[reflect.Type][]es.EventHandler {
	return nil
}

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

	streamCfg := &gcppubsub.Config{
		ProjectId: "nordic-gaming",
		TopicId:   "test_topic",
	}
	if err := gcppubsub.Reset(streamCfg); err != nil {
		t.Error(err)
		return
	}

	cli, err := gcppubsub.Open(streamCfg)
	if err != nil {
		t.Error(err)
		return
	}

	streamer, err := gpub.NewStreamer(cli)
	if err != nil {
		t.Error(err)
		return
	}

	cfg := &MyConfig{}

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

	if err := streamer.Start(ctx, cfg, callback); err != nil {
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
