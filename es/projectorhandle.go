package es

import (
	"context"
	"fmt"
	"reflect"
)

type ProjectorHandle struct {
	MethodName    string
	AggregateType reflect.Type
	EventType     reflect.Type
	Fn            reflect.Value
}

func (h *ProjectorHandle) Handle(agg interface{}, ctx context.Context, entity Entity, evt *Event) error {
	values := []reflect.Value{
		reflect.ValueOf(agg),
		reflect.ValueOf(ctx),
		reflect.ValueOf(entity),
		reflect.ValueOf(evt.Data),
	}
	out := h.Fn.Call(values)
	if len(out) != 1 {
		return fmt.Errorf("unknown error")
	}
	var err error
	if v := out[0].Interface(); v != nil {
		err = v.(error)
	}
	return err
}

func NewProjectorHandle(m reflect.Method) (*ProjectorHandle, bool) {
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
	in4 := m.Type.In(3)
	out1 := m.Type.Out(0)
	if !out1.ConvertibleTo(errType) {
		return nil, false
	}

	return &ProjectorHandle{
		MethodName:    m.Name,
		AggregateType: in3,
		EventType:     in4,
		Fn:            m.Func,
	}, true
}

func FindProjectorHandles(p interface{}) []*ProjectorHandle {
	var handles []*ProjectorHandle

	t := reflect.TypeOf(p)
	for i := 0; i < t.NumMethod(); i++ {
		method := t.Method(i)
		h, ok := NewProjectorHandle(method)
		if !ok {
			continue
		}
		handles = append(handles, h)
	}
	return handles
}
