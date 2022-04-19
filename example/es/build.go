package es

import (
	"context"
	"fmt"
	"reflect"
)

var ErrNotCommandHandler = fmt.Errorf("not a command handler")

var (
	ctxType = reflect.TypeOf((*context.Context)(nil)).Elem()
	cmdType = reflect.TypeOf((*Command)(nil)).Elem()
	errType = reflect.TypeOf((*error)(nil)).Elem()
)

type commandHandle struct {
	methodName  string
	commandType reflect.Type
	fn          reflect.Value
}

func (h *commandHandle) Handle(agg interface{}, ctx context.Context, cmd Command) error {
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

func handle(m reflect.Method) (*commandHandle, bool) {
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

	return &commandHandle{
		methodName:  m.Name,
		commandType: in3,
		fn:          m.Func,
	}, true
}

func NewCommandHandler(h interface{}) (CommandHandler, error) {
	switch impl := h.(type) {
	case Aggregate:
		return NewBaseAggregateHandler(impl)
	default:
		return nil, ErrNotCommandHandler
	}
}

type baseAggregateHandler struct {
	factory func() interface{}
	handles map[reflect.Type]*commandHandle
}

func (b *baseAggregateHandler) load(ctx context.Context, id string) (interface{}, error) {
	agg := b.factory()
	return agg, nil
}

func (b *baseAggregateHandler) Handle(ctx context.Context, cmd Command) error {
	t := reflect.TypeOf(cmd)
	h, ok := b.handles[t]
	if !ok {
		return ErrNotCommandHandler
	}

	agg, err := b.load(ctx, cmd.GetAggregateId())
	if err != nil {
		return err
	}

	// todo load it!.
	return h.Handle(agg, ctx, cmd)
}

func NewBaseAggregateHandler(agg Aggregate) (CommandHandler, error) {
	handles := map[reflect.Type]*commandHandle{}
	t := reflect.TypeOf(agg)
	for i := 0; i < t.NumMethod(); i++ {
		method := t.Method(i)
		h, ok := handle(method)
		if !ok {
			continue
		}
		handles[h.commandType] = h
	}

	raw := t
	if raw.Kind() == reflect.Ptr {
		raw = raw.Elem()
	}

	factory := func() interface{} {
		return reflect.New(raw).Interface()
	}

	return &baseAggregateHandler{
		factory: factory,
		handles: handles,
	}, nil
}
