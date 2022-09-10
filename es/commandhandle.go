package es

import (
	"context"
	"fmt"
	"reflect"

	"go.opentelemetry.io/otel"
)

var (
	ctxType = reflect.TypeOf((*context.Context)(nil)).Elem()
	cmdType = reflect.TypeOf((*Command)(nil)).Elem()
	errType = reflect.TypeOf((*error)(nil)).Elem()
)

type CommandHandle struct {
	methodName  string
	commandType reflect.Type
	fn          reflect.Value
}

func (h *CommandHandle) Handle(agg interface{}, ctx context.Context, cmd Command) error {
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

func NewCommandHandle(m reflect.Method) (*CommandHandle, bool) {
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

	return &CommandHandle{
		methodName:  m.Name,
		commandType: in3,
		fn:          m.Func,
	}, true
}

type CommandHandles map[reflect.Type]*CommandHandle

func (h CommandHandles) Handle(agg interface{}, ctx context.Context, cmd Command) error {
	pctx, pspan := otel.Tracer("CommandHandles").Start(ctx, "Handle")
	defer pspan.End()

	t := reflect.TypeOf(cmd)
	handle, ok := h[t]
	if !ok {
		return fmt.Errorf("unknown command: %s", t)
	}
	return handle.Handle(agg, pctx, cmd)
}

func NewCommandHandles(agg interface{}) CommandHandles {
	t := reflect.TypeOf(agg)
	handles := make(CommandHandles)
	for i := 0; i < t.NumMethod(); i++ {
		method := t.Method(i)
		h, ok := NewCommandHandle(method)
		if !ok {
			continue
		}
		handles[h.commandType] = h
	}
	return handles
}
