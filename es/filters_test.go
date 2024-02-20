package es

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/google/uuid"
)

type TestFilters struct {
	Name         *string      `where:"name,eq"`
	HasName      *bool        `where:"name,not.is.null"`
	Ids          *[]uuid.UUID `where:"id,in"`
	CreatedAt    *string      `order:"created_at,desc"`
	NameOrder    *string      `order:"name"`
	VersionOrder *string      `order:"string_to_array(version\\, '.')::int[]"`
}

func Test_Where(t *testing.T) {
	factory, err := NewWhereFactory[*TestFilters]()
	if err != nil {
		t.Errorf("err: %v", err)
		return
	}

	ids := []uuid.UUID{
		uuid.MustParse("702aec24-b544-4bb3-b886-04b204dd8e79"),
	}
	noIds := []uuid.UUID{}

	var yes = true
	var no = false
	var foo = "foo"
	var foo2 = "foo2"
	data := []struct {
		obj        *TestFilters
		whereCount int
		hasName    bool
		name       string
		ids        []uuid.UUID
	}{
		{
			obj:        &TestFilters{},
			whereCount: 0,
		},
		{
			obj: &TestFilters{
				Ids: &noIds,
			},
			whereCount: 1,
			ids:        noIds,
		},
		{
			obj: &TestFilters{
				Ids: &ids,
			},
			whereCount: 1,
			ids:        ids,
		},
		{
			obj: &TestFilters{
				Name: &foo,
			},
			whereCount: 1,
			name:       foo,
		},
		{
			obj: &TestFilters{
				HasName: &no,
				Name:    &foo2,
			},
			whereCount: 2,
			hasName:    false,
			name:       foo2,
		},
		{
			obj: &TestFilters{
				HasName: &yes,
				Name:    &foo2,
			},
			whereCount: 2,
			hasName:    yes,
			name:       foo2,
		},
	}

	for i, d := range data {
		t.Run(fmt.Sprintf("data[%d]", i), func(t *testing.T) {
			where := factory(d.obj)
			if where == nil {
				t.Errorf("where is nil")
			}
			clauses, ok := where.([]WhereClause)
			if !ok {
				t.Errorf("where is not []WhereClause")
				return
			}
			if len(clauses) != d.whereCount {
				t.Errorf("len(clauses) != %d", d.whereCount)
				return
			}
			for _, clause := range clauses {
				if clause.Column == "name" && clause.Op == OpEqual && clause.Args == d.name {
					continue
				}
				if clause.Column == "name" && clause.Op == OpNotIsNull && d.hasName {
					continue
				}
				if clause.Column == "name" && clause.Op == OpIsNull && !d.hasName {
					continue
				}
				if clause.Column == "id" && clause.Op == OpIn && reflect.DeepEqual(clause.Args, d.ids) {
					continue
				}

				t.Errorf("issue with column %s op %s args %v", clause.Column, clause.Op, clause.Args)
			}
		})
	}
}

func Test_Order(t *testing.T) {
	factory, err := NewOrderFactory[*TestFilters]()
	if err != nil {
		t.Errorf("err: %v", err)
		return
	}

	var nameOrder = "asc"
	data := []struct {
		obj            *TestFilters
		orderCount     int
		nameOrder      *string
		createdAtOrder string
		versionOrder   *string
	}{
		{
			obj: &TestFilters{
				NameOrder:    &nameOrder,
				VersionOrder: &nameOrder,
			},
			orderCount:     3,
			nameOrder:      &nameOrder,
			createdAtOrder: "desc",
			versionOrder:   &nameOrder,
		},
	}

	for i, d := range data {
		t.Run(fmt.Sprintf("data[%d]", i), func(t *testing.T) {
			order := factory(d.obj)
			if order == nil {
				t.Errorf("order is nil")
				return
			}
			if len(order) != d.orderCount {
				t.Errorf("len(order) != %d", d.orderCount)
				return
			}
			for _, order := range order {
				if order.Expression == "name" && d.nameOrder != nil && order.Direction == OrderDirection(*d.nameOrder) {
					continue
				}
				if order.Expression == "created_at" && order.Direction == OrderDirection(d.createdAtOrder) {
					continue
				}
				if order.Expression == "string_to_array(version, '.')::int[]" && d.versionOrder != nil && order.Direction == OrderDirection(*d.versionOrder) {
					continue
				}

				t.Errorf("issue with order %s direction %s", order.Expression, order.Direction)
			}
		})
	}
}
