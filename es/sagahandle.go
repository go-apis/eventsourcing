package es

import (
	"context"
	"fmt"
	"reflect"

	"github.com/contextcloud/eventstore/es/utils"
)

var (
	cmdsType = reflect.TypeOf(([]Command)(nil)).Elem()
)

type SagaHandle struct {
	methodName string
	eventType  reflect.Type
	fn         reflect.Value
}

func (h *SagaHandle) Handle(agg interface{}, ctx context.Context, evt Event) ([]Command, error) {
	values := []reflect.Value{
		reflect.ValueOf(agg),
		reflect.ValueOf(ctx),
		reflect.ValueOf(evt),
		reflect.ValueOf(evt.Data),
	}
	out := h.fn.Call(values)
	if len(out) != 2 {
		return nil, fmt.Errorf("unknown error")
	}
	var err error
	if v := out[1].Interface(); v != nil {
		err = v.(error)
	}
	cmds := out[0].Interface().([]Command)
	return cmds, err
}

func NewSagaHandle(m reflect.Method) (*SagaHandle, bool) {
	if !m.IsExported() {
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
	if !in3.ConvertibleTo(eventType) {
		return nil, false
	}
	in4 := m.Type.In(3)
	out1 := m.Type.Out(1)
	if !out1.ConvertibleTo(cmdsType) {
		return nil, false
	}
	out2 := m.Type.Out(2)
	if !out2.ConvertibleTo(errType) {
		return nil, false
	}

	return &SagaHandle{
		methodName: m.Name,
		eventType:  in4,
		fn:         m.Func,
	}, true
}

type SagaHandles map[reflect.Type]*SagaHandle

func (h SagaHandles) Handle(agg interface{}, ctx context.Context, evt Event) ([]Command, error) {
	t := utils.GetElemType(evt.Data)
	handle, ok := h[t]
	if !ok {
		return nil, fmt.Errorf("unknown event: %s", t)
	}
	return handle.Handle(agg, ctx, evt)
}

func NewSagaHandles(t reflect.Type) SagaHandles {
	handles := make(SagaHandles)
	for i := 0; i < t.NumMethod(); i++ {
		method := t.Method(i)
		h, ok := NewSagaHandle(method)
		if !ok {
			continue
		}
		handles[h.eventType] = h
	}
	return handles
}
