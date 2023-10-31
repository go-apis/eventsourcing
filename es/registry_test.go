package es

import "testing"

func TestEventRegistry(t *testing.T) {
	thisService := "test"

	type FakeEvent struct {
		BaseEvent `es:"myevent;service=test2;publish;alias=somethinghere,anotherhere"`
	}
	type Fake2Event struct {
		BaseEvent `es:"myevent;service=test2;publish;alias=somethinghere,anotherhere"`
	}

	evtConfig := NewEventConfig(thisService, &FakeEvent{})
	evtConfig2 := NewEventConfig(thisService, &Fake2Event{})

	t.Run("Should_Fail_AddEvent", func(t *testing.T) {
		registry := NewRegistry()
		if err := registry.AddEvent(evtConfig); err != nil {
			t.Errorf("expected no error")
		}
		if err := registry.AddEvent(evtConfig2); err == nil {
			t.Errorf("expected error")
		}
	})

	t.Run("Should_Fail_GetEvent", func(t *testing.T) {
		registry := NewRegistry()
		if err := registry.AddEvent(evtConfig); err != nil {
			t.Errorf("expected no error")
		}

		if _, err := registry.GetEventConfig(thisService, "myevent"); err == nil {
			t.Errorf("expected error")
		}
	})

	t.Run("Should_GetEvent", func(t *testing.T) {
		registry := NewRegistry()
		if err := registry.AddEvent(evtConfig); err != nil {
			t.Errorf("expected no error")
		}

		if _, err := registry.GetEventConfig("test2", "myevent"); err != nil {
			t.Errorf("expected no error")
		}
		if _, err := registry.GetEventConfig("test2", "somethinghere"); err != nil {
			t.Errorf("expected no error")
		}
		if _, err := registry.GetEventConfig("test2", "anotherhere"); err != nil {
			t.Errorf("expected no error")
		}
	})
}
