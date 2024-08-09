package es

import (
	"errors"
	"reflect"

	"github.com/go-apis/eventsourcing/es/utils"
)

var (
	boolType      = reflect.TypeOf(true)
	directionType = reflect.TypeOf("")
)

var ErrInvalidTag = errors.New(`invalid tag`)

type Filter struct {
	Distinct []interface{}
	Where    Where
	Order    []Order
	Limit    *int
	Offset   *int
}

type OrderDirection string

const (
	OrderAsc  OrderDirection = `asc`
	OrderDesc OrderDirection = `desc`
)

type Order struct {
	Expression string
	Direction  OrderDirection
}

type Where interface{}

type WhereOr struct {
	Where
}

type WhereClause struct {
	Column string
	Op     Op
	Args   interface{}
}

func Limit(v int) *int {
	return &v
}
func Offset(v int) *int {
	return &v
}

type Op string

const (
	OpEqual    Op = `eq`
	OpNotEqual Op = `not.eq`

	OpGreaterThan    Op = `gt`
	OpGreaterOrEqual Op = `gte`
	OpLessThan       Op = `lt`
	OpLessOrEqual    Op = `lte`
	OpLike           Op = `like`
	OpNotLike        Op = `not.like`
	OpIs             Op = `is`
	OpNotIs          Op = `not.is`
	OpIsNull         Op = `is.null`
	OpNotIsNull      Op = `not.is.null`
	OpIn             Op = `in`
	OpNotIn          Op = `not.in`
)

func InverseOp(op Op) Op {
	switch op {
	case OpEqual:
		return OpNotEqual
	case OpNotEqual:
		return OpEqual
	case OpGreaterThan:
		return OpLessOrEqual
	case OpGreaterOrEqual:
		return OpLessThan
	case OpLessThan:
		return OpGreaterOrEqual
	case OpLessOrEqual:
		return OpGreaterThan
	case OpLike:
		return OpNotLike
	case OpNotLike:
		return OpLike
	case OpIs:
		return OpNotIs
	case OpNotIs:
		return OpIs
	case OpIsNull:
		return OpNotIsNull
	case OpNotIsNull:
		return OpIsNull
	case OpIn:
		return OpNotIn
	case OpNotIn:
		return OpIn
	default:
		panic(`unknown op`)
	}
}

func IsBoolOp(op Op) bool {
	switch op {
	case OpIs, OpNotIs, OpIsNull, OpNotIsNull:
		return true
	}
	return false
}

func IsBoolType(t reflect.Type) bool {
	ref := t
	if t.Kind() == reflect.Ptr {
		ref = t.Elem()
	}
	return ref == boolType
}

func IsDirection(t reflect.Type) bool {
	ref := t
	if t.Kind() == reflect.Ptr {
		ref = t.Elem()
	}
	return ref == directionType
}

type WhereHandle[T any] struct {
	FieldName string
	Col       string
	Op        Op
}

func (w WhereHandle[T]) Resolve(obj T) *WhereClause {
	t := reflect.ValueOf(obj)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	field := t.FieldByName(w.FieldName)
	for field.Kind() == reflect.Ptr {
		if field.IsNil() {
			return nil
		}

		field = field.Elem()
	}

	if IsBoolOp(w.Op) {
		op := w.Op
		val := field.Bool()

		if !val {
			op = InverseOp(op)
		}

		return &WhereClause{
			Column: w.Col,
			Op:     op,
		}
	}

	val := field.Interface()
	return &WhereClause{
		Column: w.Col,
		Op:     w.Op,
		Args:   val,
	}
}

type WhereFactory[T any] func(T) Where

func NewWhereFactory[T any]() (WhereFactory[T], error) {
	var obj T
	t := reflect.TypeOf(obj)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	var handles []WhereHandle[T]

	count := t.NumField()
	for i := 0; i < count; i++ {
		field := t.Field(i)
		tag := field.Tag.Get(`where`)
		if tag == `` {
			continue
		}

		fieldName := field.Name
		name := field.Name
		op := OpEqual

		// do stuff.
		split := utils.SplitTag(tag)
		switch len(split) {
		case 1:
			op = Op(split[0])
		case 2:
			name = split[0]
			op = Op(split[1])
		default:
			return nil, ErrInvalidTag
		}

		// validate it.
		if IsBoolOp(op) && !IsBoolType(field.Type) {
			return nil, ErrInvalidTag
		}

		handles = append(handles, WhereHandle[T]{
			FieldName: fieldName,
			Col:       name,
			Op:        op,
		})
	}

	return func(obj T) Where {
		var columns []WhereClause
		for _, handle := range handles {
			clause := handle.Resolve(obj)
			if clause != nil {
				columns = append(columns, *clause)
			}
		}
		return columns
	}, nil
}

type OrderHandle[T any] struct {
	FieldName  string
	Expression string
	Direction  *OrderDirection
}

func (w OrderHandle[T]) value(obj T) *string {
	t := reflect.ValueOf(obj)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	field := t.FieldByName(w.FieldName)
	for field.Kind() == reflect.Ptr {
		if field.IsNil() {
			return nil
		}

		field = field.Elem()
	}

	str := field.String()
	return &str
}

func (w OrderHandle[T]) Resolve(obj T) *Order {
	value := w.value(obj)
	if value == nil && w.Direction == nil {
		return nil
	}

	var direction OrderDirection
	if w.Direction != nil {
		direction = *w.Direction
	}
	if value != nil {
		v := *value
		direction = OrderDirection(v)
	}

	return &Order{
		Expression: w.Expression,
		Direction:  direction,
	}
}

type OrderFactory[T any] func(T) []Order

func NewOrderFactory[T any]() (OrderFactory[T], error) {
	var obj T
	t := reflect.TypeOf(obj)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	var handles []OrderHandle[T]

	count := t.NumField()
	for i := 0; i < count; i++ {
		field := t.Field(i)
		tag := field.Tag.Get(`order`)
		if tag == `` {
			continue
		}

		fieldName := field.Name
		name := field.Name
		var direction *OrderDirection

		// do stuff.
		split := utils.SplitTag(tag)
		switch len(split) {
		case 1:
			name = split[0]
		case 2:
			d := OrderDirection(split[1])

			name = split[0]
			direction = &d
		default:
			return nil, ErrInvalidTag
		}

		// validate it.
		if !IsDirection(field.Type) {
			return nil, ErrInvalidTag
		}

		handles = append(handles, OrderHandle[T]{
			FieldName:  fieldName,
			Expression: name,
			Direction:  direction,
		})
	}

	return func(obj T) []Order {
		var orders []Order
		for _, handle := range handles {
			order := handle.Resolve(obj)
			if order != nil {
				orders = append(orders, *order)
			}
		}
		return orders
	}, nil
}
