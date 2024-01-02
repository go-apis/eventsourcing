package es

import (
	"context"
	"fmt"
	"reflect"
)

type EventHandlerHandle struct {
	methodName string
	eventType  reflect.Type
	fn         reflect.Value
}

func (h *EventHandlerHandle) Handle(agg interface{}, ctx context.Context, evt *Event) error {
	values := []reflect.Value{
		reflect.ValueOf(agg),
		reflect.ValueOf(ctx),
		reflect.ValueOf(evt),
		reflect.ValueOf(evt.Data),
	}
	out := h.fn.Call(values)
	if len(out) != 1 {
		return fmt.Errorf("unknown error")
	}
	var err error
	if v := out[0].Interface(); v != nil {
		err = v.(error)
	}
	return err
}

func NewEventHandlerHandle(m reflect.Method) (*EventHandlerHandle, bool) {
	if !m.IsExported() {
		return nil, false
	}
	if m.Name == "Run" {
		return nil, false
	}

	numIn := m.Type.NumIn()
	if numIn != 4 {
		return nil, false
	}
	numOut := m.Type.NumOut()
	if numOut != 1 {
		return nil, false
	}

	in2 := m.Type.In(1)
	if !in2.ConvertibleTo(ctxType) {
		return nil, false
	}
	in3 := m.Type.In(2)
	if in3.Kind() != reflect.Ptr || !in3.Elem().ConvertibleTo(eventType) {
		return nil, false
	}
	in4 := m.Type.In(3)
	out1 := m.Type.Out(0)
	if !out1.ConvertibleTo(errType) {
		return nil, false
	}

	return &EventHandlerHandle{
		methodName: m.Name,
		eventType:  in4,
		fn:         m.Func,
	}, true
}

type EventHandlerHandles map[reflect.Type]*EventHandlerHandle

func (h EventHandlerHandles) Handle(agg interface{}, ctx context.Context, evt *Event) error {
	t := reflect.TypeOf(evt.Data)
	handle, ok := h[t]
	if !ok {
		return fmt.Errorf("unknown event: %s", t)
	}
	return handle.Handle(agg, ctx, evt)
}

func NewEventHandlerHandles(s IsEventHandler) EventHandlerHandles {
	t := reflect.TypeOf(s)
	handles := make(EventHandlerHandles)
	for i := 0; i < t.NumMethod(); i++ {
		method := t.Method(i)
		h, ok := NewEventHandlerHandle(method)
		if !ok {
			continue
		}
		handles[h.eventType] = h
	}
	return handles
}
