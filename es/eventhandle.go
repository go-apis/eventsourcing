package es

import (
	"context"
	"fmt"
	"reflect"
)

var (
	eventType = reflect.TypeOf((*Event)(nil)).Elem()
)

type EventHandle struct {
	methodName string
	eventType  reflect.Type
	fn         reflect.Value
}

func (h *EventHandle) Handle(agg interface{}, ctx context.Context, cmd Command) error {
	values := []reflect.Value{
		reflect.ValueOf(agg),
		reflect.ValueOf(ctx),
		reflect.ValueOf(cmd),
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

func NewEventHandle(m reflect.Method) (*EventHandle, bool) {
	if m.Name == "Apply" {
		return nil, false
	}
	if !m.IsExported() {
		return nil, false
	}

	numIn := m.Type.NumIn()
	if numIn != 3 {
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
	if !in3.ConvertibleTo(cmdType) {
		return nil, false
	}
	out1 := m.Type.Out(0)
	if !out1.ConvertibleTo(errType) {
		return nil, false
	}

	return &EventHandle{
		methodName: m.Name,
		eventType:  in3,
		fn:         m.Func,
	}, true
}

type EventHandles map[reflect.Type]*EventHandle

func NewEventHandles(t reflect.Type) EventHandles {
	handles := make(EventHandles)
	for i := 0; i < t.NumMethod(); i++ {
		method := t.Method(i)
		h, ok := NewEventHandle(method)
		if !ok {
			continue
		}
		handles[h.eventType] = h
	}
	return handles
}
