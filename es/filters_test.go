package es

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/google/uuid"
)

type TestWhere struct {
	Name    *string      `where:"name,eq"`
	HasName *bool        `where:"name,not.is.null"`
	Ids     *[]uuid.UUID `where:"id,in"`
}

func Test_It(t *testing.T) {
	factory, err := NewWhereFactory[*TestWhere]()
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
		obj        *TestWhere
		whereCount int
		hasName    bool
		name       string
		ids        []uuid.UUID
	}{
		{
			obj:        &TestWhere{},
			whereCount: 0,
		},
		{
			obj: &TestWhere{
				Ids: &noIds,
			},
			whereCount: 1,
			ids:        noIds,
		},
		{
			obj: &TestWhere{
				Ids: &ids,
			},
			whereCount: 1,
			ids:        ids,
		},
		{
			obj: &TestWhere{
				Name: &foo,
			},
			whereCount: 1,
			name:       foo,
		},
		{
			obj: &TestWhere{
				HasName: &no,
				Name:    &foo2,
			},
			whereCount: 2,
			hasName:    false,
			name:       foo2,
		},
		{
			obj: &TestWhere{
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
